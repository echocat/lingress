package rules

import (
	"fmt"
	"github.com/echocat/lingress/definition"
	"github.com/echocat/lingress/kubernetes"
	"github.com/echocat/lingress/support"
	"github.com/echocat/lingress/value"
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
	Host value.Fqdn
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
	HostWildcard value.ForcibleString

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
			HostWildcard: value.NewForcibleString("", false),

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
	fe.Flag("hosts.wildcard", "Default wildcard pattern for ingress configuration."+
		" As Kubernetes ingress configuration does not support wildcards as '*' this argument specifies"+
		" a specific domain segment to be a placeholder; like 'x--wildcard--x' instead of '*'."+
		" If empty no wildcards are supported at all."+
		" This can be overwritten by the 'lingress.echocat.org/hosts-wildcard' annotation."+
		" If this parameter is prefixed with '!' the annotation cannot override this behavior.").
		PlaceHolder(instance.HostWildcard.String()).
		Envar(support.FlagEnvName(appPrefix, "HOSTS_WILDCARD")).
		SetValue(&instance.HostWildcard)

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

	if v := ingress.Spec.Backend; v != nil {
		if backend, err := instance.ingressToBackend(source, *v); err != nil {
			return err
		} else if options, err := instance.newOptionsBy(ingress); err != nil {
			return err
		} else if backend != nil {
			r := NewRule("", []string{}, source, backend, options)
			if err := target.Put(r); err != nil {
				return err
			}
		}
	}

	for _, forHost := range ingress.Spec.Rules {
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
				} else if options, err := instance.newOptionsBy(ingress); err != nil {
					return err
				} else if host, err := instance.parseHost(&forHost, options); err != nil {
					log.WithField("service", fmt.Sprintf("%s/%s", source.namespace, forPath.Backend.ServiceName)).
						WithField("port", forPath.Backend.ServicePort).
						WithField("source", source.String()).
						WithField("path", forPath.Path).
						WithError(err).
						Warn("illegal host in ingress; ingress will not functioning")
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

func (instance *repositoryImplState) parseHost(ingressRule *v1beta1.IngressRule, options Options) (value.WildcardSupportingFqdn, error) {
	plain := normalizeHostname(ingressRule.Host)
	var hostWildcardFromOptions value.String
	if v, ok := options[optionsMatchingKey].(*OptionsMatching); ok {
		hostWildcardFromOptions = v.HostWildcard
	}
	hostWildcard := instance.HostWildcard.Evaluate(hostWildcardFromOptions, "")
	if hostWildcard != "" {
		if strings.HasPrefix(plain, hostWildcard.String()+".") {
			plain = "*" + plain[len(hostWildcard):]
		}
	}
	var result value.WildcardSupportingFqdn
	if err := result.Set(plain); err != nil {
		return "", err
	}
	return result, nil
}

func (instance *repositoryImplState) newOptionsBy(ingress *v1beta1.Ingress) (Options, error) {
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

	service, err := instance.ingressToService(source, ib)
	if err != nil {
		return nil, err
	} else if service == nil {
		l.Warn("service not found; maybe orphan ingress?; ingress will not functioning")
		return nil, nil
	}
	if service.Spec.Type != v1.ServiceTypeClusterIP {
		l.WithField("serviceType", service.Spec.Type).
			Warn("unsupported serviceType; ingress will not functioning")
		return nil, nil
	}
	if strings.TrimSpace(service.Spec.ClusterIP) == "" {
		l.Warnf("serviceType is '%s' but clusterIP of service is not set; ingress will not functioning", v1.ServiceTypeClusterIP)
		return nil, nil
	}
	port, err := instance.evaluateServicePort(ib.ServicePort, service)
	if err != nil {
		l.WithError(err).
			Warn("cannot resolve backend port; ingress will not functioning")
		return nil, nil
	}
	addr, err := instance.clusterIpBasedServiceToAddr(service.Spec.ClusterIP, port)
	if err != nil {
		l.WithError(err).
			Warn("cannot resolve backend address; ingress will not functioning")
		return nil, nil
	}

	return addr, nil
}

func (instance *repositoryImplState) evaluateServicePort(in intstr.IntOrString, service *v1.Service) (int32, error) {
	switch in.Type {
	case intstr.Int:
		return in.IntVal, nil
	case intstr.String:
		for _, candidate := range service.Spec.Ports {
			if candidate.Name == in.StrVal {
				return candidate.Port, nil
			}
		}
		return 0, fmt.Errorf("unknown service reference %s:%s", service.Name, in.StrVal)
	default:
		return 0, fmt.Errorf("only port type String or Int is currently supported; but got: %d", in.Type)
	}
}

func (instance *repositoryImplState) clusterIpBasedServiceToAddr(ipStr string, port int32) (net.Addr, error) {
	if ips, err := net.LookupIP(ipStr); err != nil {
		return nil, err
	} else if len(ips) <= 0 {
		return nil, fmt.Errorf("host %s cannot be resovled to any IP address", ipStr)
	} else {
		return &net.TCPAddr{
			IP:   ips[0],
			Port: int(port),
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
	host := q.Host
	path, err := ParsePath(q.Path, true)
	if err != nil {
		return nil, err
	}
	return instance.ByHostRules.Find(host, path)
}

func (instance *KubernetesBasedRepository) FindCertificatesBy(q CertificateQuery) (Certificates, error) {
	return instance.CertificatesByHost.Find(q.Host), nil
}
