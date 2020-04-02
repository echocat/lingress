package rules

import (
	"fmt"
	"github.com/echocat/lingress/definition"
	"github.com/echocat/lingress/kubernetes"
	"github.com/echocat/lingress/support"
	log "github.com/sirupsen/logrus"
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

type CertificateQuery struct {
	Host string
}

type Repository interface {
	support.FlagRegistrar
	Init(stop support.Channel) error
	All(consumer func(Rule) error) error
	FindBy(Query) (Rules, error)
}

type CombinedRepository interface {
	Repository
	CertificateRepository
}

type KubernetesBasedRepository struct {
	Environment  *kubernetes.Environment
	ByHostRules  *ByHost
	ResyncAfter  time.Duration
	IngressClass []string

	CertificatesSecret string
	CertificatesByHost CertificatesByHost
	OptionsFactory     OptionsFactory
}

func NewRepository() (CombinedRepository, error) {
	if environment, err := kubernetes.NewEnvironment(); err != nil {
		return nil, err
	} else {
		result := &KubernetesBasedRepository{
			Environment:  environment,
			IngressClass: []string{},

			CertificatesSecret: "certificates",
			OptionsFactory:     DefaultOptionsFactory,
		}
		result.ByHostRules = NewByHost(result.onRuleAdded, result.onRuleRemoved)
		return result, nil
	}
}

func (instance *KubernetesBasedRepository) RegisterFlag(fe support.FlagEnabled, appPrefix string) error {
	fe.Flag("resyncAfter", "Time after which the configuration should be resynced to be ensure to be not out of date.").
		PlaceHolder(instance.ResyncAfter.String()).
		Envar(support.FlagEnvName(appPrefix, "RESYNC_AFTER")).
		DurationVar(&instance.ResyncAfter)
	fe.Flag("ingressClass", "Ingress classes which this application should respect.").
		PlaceHolder(ingressClass).
		Envar(support.FlagEnvName(appPrefix, "INGRESS_CLASS")).
		StringsVar(&instance.IngressClass)
	fe.Flag("secret.certificates", "Name of the secret that contains the certificates.").
		PlaceHolder(instance.CertificatesSecret).
		Envar(support.FlagEnvName(appPrefix, "SECRET_CERTIFICATES")).
		StringVar(&instance.CertificatesSecret)

	return instance.Environment.RegisterFlag(fe, appPrefix)
}

func (instance *KubernetesBasedRepository) Init(stop support.Channel) error {
	if len(instance.IngressClass) == 0 {
		instance.IngressClass = []string{ingressClass, ""}
	}
	log.Info("initial sync of definitions...")

	client, err := instance.Environment.NewClient()
	if err != nil {
		return err
	}
	definitions, err := definition.New(client, instance.ResyncAfter)
	if err != nil {
		return err
	}
	definitions.SetNamespace(instance.Environment.Namespace)

	state := &repositoryImplState{
		KubernetesBasedRepository: instance,
		definitions:               definitions,
	}

	state.initiated.Store(false)

	definitions.Ingress.OnElementAdded = state.onIngressElementAdded
	definitions.Ingress.OnElementUpdated = state.onIngressElementUpdated
	definitions.Ingress.OnElementRemoved = state.onIngressElementRemoved

	definitions.ServiceSecrets.OnElementAdded = state.onServiceSecretsElementAdded
	definitions.ServiceSecrets.OnElementUpdated = state.onServiceSecretsElementUpdated
	definitions.ServiceSecrets.OnElementRemoved = state.onServiceSecretsElementRemoved

	if err := definitions.Init(stop); err != nil {
		return err
	}

	state.initiated.Store(true)

	log.Info("initial sync of definitions... done!")
	return nil
}

func (instance *KubernetesBasedRepository) onRuleAdded(_ []string, r Rule) {
	log.WithField("rule", r).Debug("rule added")
}

func (instance *KubernetesBasedRepository) onRuleRemoved(_ []string, r Rule) {
	log.WithField("rule", r).Debug("rule removed")
}

type repositoryImplState struct {
	*KubernetesBasedRepository

	definitions *definition.Definitions
	initiated   atomic.Value
}

func (instance *repositoryImplState) onSecretCertificatesChanged(key string, new metav1.Object) error {
	if new == nil {
		instance.CertificatesByHost = CertificatesByHost{}
		return nil
	}

	s := new.(*v1.Secret)

	cbh := CertificatesByHost{}

	for file, candidate := range s.Data {
		base, ext := support.SplitExt(file)
		if ext == ".crt" || ext == ".cer" {
			privateKeyFile := base + ".key"
			if pk, ok := s.Data[privateKeyFile]; !ok {
				log.WithField("secret", key).
					WithField("certificate", file).
					WithField("privateKey", privateKeyFile).
					Warn("cannot find expected privateKey in secret for provided certificate; ignoring...")
			} else if err := cbh.AddBytes(candidate, pk); err != nil {
				log.WithError(err).
					WithField("secret", key).
					WithField("certificate", file).
					WithField("privateKey", privateKeyFile).
					Warn("cannot parse certificate and privateKey pair from secret; ignoring...")
			}
		}
	}

	instance.CertificatesByHost = cbh

	return nil
}

func (instance *repositoryImplState) expectedCertificatesKey() string {
	return instance.Environment.Namespace + "/" + instance.CertificatesSecret
}

func (instance *repositoryImplState) onServiceSecretsElementAdded(key string, new metav1.Object) error {
	if key == instance.expectedCertificatesKey() {
		return instance.onSecretCertificatesChanged(key, new)
	}
	return nil
}

func (instance *repositoryImplState) onServiceSecretsElementUpdated(key string, _, new metav1.Object) error {
	if key == instance.expectedCertificatesKey() {
		return instance.onSecretCertificatesChanged(key, new)
	}
	return nil
}

func (instance *repositoryImplState) onServiceSecretsElementRemoved(key string) error {
	if key == instance.expectedCertificatesKey() {
		return instance.onSecretCertificatesChanged(key, nil)
	}
	return nil
}

func (instance *repositoryImplState) onIngressElementAdded(_ string, new metav1.Object) error {
	target := instance.ByHostRules
	clonedUpdate := instance.initiated.Load() == true
	if clonedUpdate {
		target = target.Clone()
	}

	candidate := new.(*v1beta1.Ingress)

	annotations := candidate.GetAnnotations()
	ic := annotations[annotationIngressClass]
	ingressClassMatches := false
	for _, candidate := range instance.IngressClass {
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
		instance.ByHostRules = target
	}
	return nil
}

func (instance *repositoryImplState) onIngressElementUpdated(key string, _, new metav1.Object) error {
	candidate := new.(*v1beta1.Ingress)

	annotations := candidate.GetAnnotations()
	ic := annotations[annotationIngressClass]
	for _, candidate := range instance.IngressClass {
		if ic == candidate {
			return instance.onIngressElementAdded(key, new)
		}
	}
	return instance.onIngressElementRemoved(key)
}

func (instance *repositoryImplState) onIngressElementRemoved(key string) error {
	target := instance.ByHostRules
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
		instance.ByHostRules = target
	}
	return nil
}

func (instance *repositoryImplState) visitIngress(ingress *v1beta1.Ingress, target *ByHost) error {
	source := &sourceReference{
		namespace: ingress.Namespace,
		name:      ingress.Name,
	}
	if err := target.Remove(PredicateBySource(source)); err != nil {
		return err
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
				} else if options, err := instance.mewOptionsBy(ingress); err != nil {
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

func (instance *repositoryImplState) mewOptionsBy(ingress *v1beta1.Ingress) (Options, error) {
	result := instance.OptionsFactory()
	if err := result.Set(ingress.GetAnnotations()); err != nil {
		return nil, err
	}

	return result, nil
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

func (instance *KubernetesBasedRepository) All(consumer func(Rule) error) error {
	return instance.ByHostRules.All(consumer)
}

func (instance *KubernetesBasedRepository) FindBy(q Query) (Rules, error) {
	host := normalizeHostname(q.Host)
	path, err := ParsePath(q.Path, true)
	if err != nil {
		return nil, err
	}
	return instance.ByHostRules.Find(host, path)
}

func (instance *KubernetesBasedRepository) FindCertificatesBy(q CertificateQuery) (Certificates, error) {
	return instance.CertificatesByHost.Find(q.Host), nil
}
