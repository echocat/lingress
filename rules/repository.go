package rules

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"github.com/echocat/lingress/definition"
	"github.com/echocat/lingress/kubernetes"
	"github.com/echocat/lingress/support"
	v1 "k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/tools/cache"
	"net"
	"strings"
	"sync/atomic"
	"time"
)

type Query struct {
	Host string
	Path string
}

type Repository interface {
	support.FlagRegistrar
	Init(stopCh chan struct{}) error
	All(consumer func(Rule) error) error
	FindBy(Query) (Rules, error)
}

type repositoryImpl struct {
	environment  *kubernetes.Environment
	byHostRules  *ByHost
	resyncAfter  time.Duration
	ingressClass []string
}

func NewRepository() (Repository, error) {
	if environment, err := kubernetes.NewEnvironment(); err != nil {
		return nil, err
	} else {
		result := &repositoryImpl{
			environment:  environment,
			ingressClass: []string{},
		}
		result.byHostRules = NewByHost(result.onRuleAdded, result.onRuleRemoved)
		return result, nil
	}
}

func (instance *repositoryImpl) RegisterFlag(fe support.FlagEnabled, appPrefix string) error {
	fe.Flag("resyncAfter", "Time after which the configuration should be resynced to be ensure to be not out of date.").
		PlaceHolder(instance.resyncAfter.String()).
		Envar(support.FlagEnvName(appPrefix, "RESYNC_AFTER")).
		DurationVar(&instance.resyncAfter)
	fe.Flag("ingressClass", "Ingress classes which this application should respect.").
		PlaceHolder(ingressClass).
		Envar(support.FlagEnvName(appPrefix, "INGRESS_CLASS")).
		StringsVar(&instance.ingressClass)

	return instance.environment.RegisterFlag(fe, appPrefix)
}

func (instance *repositoryImpl) Init(stopCh chan struct{}) error {
	if len(instance.ingressClass) == 0 {
		instance.ingressClass = []string{ingressClass, ""}
	}
	log.Info("initial sync of definitions...")

	client, err := instance.environment.NewClient()
	if err != nil {
		return err
	}
	definitions, err := definition.New(client, instance.resyncAfter)
	if err != nil {
		return err
	}

	state := &repositoryImplState{
		repositoryImpl: instance,
		definitions:    definitions,
	}

	state.initiated.Store(false)

	definitions.Ingress.OnElementAdded = state.onElementAdded
	definitions.Ingress.OnElementUpdated = state.onElementUpdated
	definitions.Ingress.OnElementRemoved = state.onElementRemoved

	if err := definitions.Init(stopCh); err != nil {
		return err
	}

	state.initiated.Store(true)

	log.Info("initial sync of definitions... done!")
	return nil
}

func (instance *repositoryImpl) onRuleAdded(path []string, r Rule) {
	log.WithField("rule", r).Info("rule added")
}

func (instance *repositoryImpl) onRuleRemoved(path []string, r Rule) {
	log.WithField("rule", r).Info("rule removed")
}

type repositoryImplState struct {
	*repositoryImpl

	definitions *definition.Definitions
	initiated   atomic.Value
}

func (instance *repositoryImplState) onElementAdded(key string, new metav1.Object) error {
	target := instance.byHostRules
	clonedUpdate := instance.initiated.Load() == true
	if clonedUpdate {
		target = target.Clone()
	}

	candidate := new.(*v1beta1.Ingress)

	annotations := candidate.GetAnnotations()
	ic := annotations[annotationIngressClass]
	ingressClassMatches := false
	for _, candidate := range instance.ingressClass {
		if ic == candidate {
			ingressClassMatches = true
			break
		}
	}

	if !ingressClassMatches {
		return nil
	}

	if err := instance.visitIngress(candidate, target); err != nil {
		return err
	}

	if clonedUpdate {
		instance.byHostRules = target
	}
	return nil
}

func (instance *repositoryImplState) onElementUpdated(key string, old, new metav1.Object) error {
	candidate := new.(*v1beta1.Ingress)

	annotations := candidate.GetAnnotations()
	ic := annotations[annotationIngressClass]
	for _, candidate := range instance.ingressClass {
		if ic == candidate {
			return instance.onElementAdded(key, new)
		}
	}
	return instance.onElementRemoved(key)
}

func (instance *repositoryImplState) onElementRemoved(key string) error {
	target := instance.byHostRules
	clonedUpdate := instance.initiated.Load() != true
	if clonedUpdate {
		target = target.Clone()
	}

	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		return fmt.Errorf("cannot parse key '%s' to namespace and name: %v", key, err)
	}

	source := sourceReference{
		namespace: namespace,
		name:      name,
	}

	if err := target.Remove(PredicateBySource(source)); err != nil {
		return fmt.Errorf("cannot remove previous element by source '%s': %v", source, err)
	}

	if clonedUpdate {
		instance.byHostRules = target
	}
	return nil
}

func (instance *repositoryImplState) visitIngress(ingress *v1beta1.Ingress, target *ByHost) error {
	source := &sourceReference{
		namespace: ingress.Namespace,
		name:      ingress.Name,
	}

	for _, forHost := range ingress.Spec.Rules {
		host := normalizeHostname(forHost.Host)
		if forHost.IngressRuleValue.HTTP != nil && forHost.IngressRuleValue.HTTP.Paths != nil {
			for _, forPath := range forHost.IngressRuleValue.HTTP.Paths {
				if path, err := ParsePath(forPath.Path, false); err != nil {
					log.WithField("service", fmt.Sprintf("%s/%s", source.namespace, forPath.Backend.ServiceName)).
						WithField("port", forPath.Backend.ServicePort).
						WithField("source", source.String()).
						WithField("path", forPath.Path).
						WithError(err).
						Warn("illegal path in ingress; ingress will not functioning")
				} else if backend, err := instance.ingressToBackend(source, forPath.Backend); err != nil {
					return err
				} else if options, err := optionsForIngress(ingress); err != nil {
					return err
				} else if backend != nil {
					r := NewRule(host, path, source, backend, options)
					if err := target.Put(r); err != nil {
						return err
					}
				}
			}
		}
	}

	return nil
}

func (instance *repositoryImplState) ingressToBackend(source *sourceReference, ib v1beta1.IngressBackend) (net.Addr, error) {
	l := log.WithField("service", ib.ServiceName).
		WithField("port", ib.ServicePort).
		WithField("source", source.String())

	if service, err := instance.ingressToService(source, ib); err != nil {
		return nil, err
	} else if service == nil {
		l.Warn("service not found; maybe orphan ingress?; ingress will not functioning")
		return nil, nil
	} else if service.Spec.Type != v1.ServiceTypeClusterIP {
		l.WithField("serviceType", service.Spec.Type).
			Warn("unsupported serviceType; ingress will not functioning")
		return nil, nil
	} else if strings.TrimSpace(service.Spec.ClusterIP) == "" {
		l.Warnf("serviceType is '%s' but clusterIP of service is not set; ingress will not functioning", v1.ServiceTypeClusterIP)
		return nil, nil
	} else if addr, err := instance.clusterIpBasedServiceToAddr(service.Spec.ClusterIP, ib.ServicePort); err != nil {
		l.WithError(err).
			Warn("cannot resolve backend address; ingress will not functioning")
		return nil, nil
	} else {
		return addr, nil
	}
}

func (instance *repositoryImplState) clusterIpBasedServiceToAddr(ipStr string, portStr intstr.IntOrString) (net.Addr, error) {
	if ips, err := net.LookupIP(ipStr); err != nil {
		return nil, err
	} else if len(ips) <= 0 {
		return nil, fmt.Errorf("host %s cannot be resovled to any IP address", ipStr)
	} else if port, err := net.LookupPort("tcp", portStr.String()); err != nil {
		return nil, fmt.Errorf("cannot handle port of type %v: %v", portStr, err)
	} else {
		return &net.TCPAddr{
			IP:   ips[0],
			Port: port,
		}, nil
	}
}

func (instance *repositoryImplState) ingressToService(source *sourceReference, ib v1beta1.IngressBackend) (*v1.Service, error) {
	serviceKey := fmt.Sprintf("%s/%s", source.namespace, ib.ServiceName)

	return instance.definitions.Service.Get(serviceKey)
}

func (instance *repositoryImpl) All(consumer func(Rule) error) error {
	return instance.byHostRules.All(consumer)
}

func (instance *repositoryImpl) FindBy(q Query) (Rules, error) {
	host := normalizeHostname(q.Host)
	path, err := ParsePath(q.Path, true)
	if err != nil {
		return nil, err
	}
	return instance.byHostRules.Find(host, path)
}
