package rules

import (
	"fmt"
	"github.com/echocat/lingress/definition"
	"github.com/echocat/lingress/kubernetes"
	"github.com/echocat/lingress/support"
	"github.com/echocat/lingress/value"
	"github.com/echocat/slf4g"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	Logger       log.Logger

	CertificatesSecrets []string
	CertificatesByHost  CertificatesByHost
	OptionsFactory      OptionsFactory
}

func NewRepository(logger log.Logger) (CombinedRepository, error) {
	if environment, err := kubernetes.NewEnvironment(); err != nil {
		return nil, err
	} else {
		result := &KubernetesBasedRepository{
			Environment:  environment,
			IngressClass: []string{},

			CertificatesSecrets: []string{"certificates"},
			OptionsFactory:      DefaultOptionsFactory,

			Logger: logger,
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
		PlaceHolder(strings.Join(instance.CertificatesSecrets, ",")).
		Envar(support.FlagEnvName(appPrefix, "SECRET_CERTIFICATES")).
		StringsVar(&instance.CertificatesSecrets)

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
	definitions, err := definition.New(client, instance.ResyncAfter, instance.Logger)
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
	instance.Logger.With("rule", r).Debug("rule added")
}

func (instance *KubernetesBasedRepository) onRuleRemoved(_ []string, r Rule) {
	instance.Logger.With("rule", r).Debug("rule removed")
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

	l := instance.Logger.
		With("key", key)

	cbh := CertificatesByHost{}

	for file, candidate := range s.Data {
		l := l.With("certificate", file)
		base, ext := support.SplitExt(file)
		if ext == ".crt" || ext == ".cer" {
			privateKeyFile := base + ".key"
			l := l.With("privateKey", privateKeyFile)
			pk, ok := s.Data[privateKeyFile]
			if !ok {
				l.Warn("cannot find expected privateKey in secret for provided certificate; ignoring...")
				continue
			}
			if err := cbh.AddBytes(candidate, pk); err != nil {
				l.WithError(err).Warn("cannot parse certificate and privateKey pair from secret; ignoring...")
				continue
			}
		}
	}

	instance.CertificatesByHost = cbh

	return nil
}

func (instance *repositoryImplState) isExpectedCertificatesKey(what string) bool {
	for _, candidate := range instance.CertificatesSecrets {
		if strings.Contains(candidate, "/") {
			candidate = instance.Environment.Namespace + "/" + candidate
		}
		if candidate == what {
			return true
		}
	}
	return false
}

func (instance *repositoryImplState) onServiceSecretsElementAdded(key string, new metav1.Object) error {
	if instance.isExpectedCertificatesKey(key) {
		return instance.onSecretCertificatesChanged(key, new)
	}
	return nil
}

func (instance *repositoryImplState) onServiceSecretsElementUpdated(key string, _, new metav1.Object) error {
	if instance.isExpectedCertificatesKey(key) {
		return instance.onSecretCertificatesChanged(key, new)
	}
	return nil
}

func (instance *repositoryImplState) onServiceSecretsElementRemoved(key string) error {
	if instance.isExpectedCertificatesKey(key) {
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

	candidate := new.(*networkingv1.Ingress)

	if !instance.matchesIngressClass(candidate) {
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
	candidate := new.(*networkingv1.Ingress)

	if instance.matchesIngressClass(candidate) {
		return instance.onIngressElementAdded(key, new)
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

func (instance *repositoryImplState) matchesIngressClass(what *networkingv1.Ingress) bool {
	requested := what.Spec.IngressClassName
	if requested == nil {
		for _, candidate := range instance.IngressClass {
			if candidate == "" || candidate == "*" {
				return true
			}
		}
		return false
	}
	for _, candidate := range instance.IngressClass {
		if candidate == *requested || candidate == "*" {
			return true
		}
	}
	return false
}

func (instance *repositoryImplState) visitIngress(ingress *networkingv1.Ingress, target *ByHost) error {
	source := &sourceReference{
		namespace: ingress.Namespace,
		name:      ingress.Name,
	}
	if err := target.Remove(PredicateBySource(source)); err != nil {
		return err
	}

	l := instance.Logger.
		With("type", "ingress").
		Withf("key", "%s/%s", source.namespace, source.name)

	if v := ingress.Spec.DefaultBackend; v != nil {
		l := l.With("kind", "defaultBackend")
		backend, err := instance.ingressToBackend(source, *v, l)
		if err != nil {
			return err
		}
		if backend != nil {
			options, err := instance.newOptionsBy(ingress)
			if err != nil {
				return err
			}
			r := NewRule("", []string{}, PathTypePrefix, source, backend, options)
			if err := target.Put(r); err != nil {
				return err
			}
			l.Debug("element registered")
		}
	}

	for _, forHost := range ingress.Spec.Rules {
		if forHost.IngressRuleValue.HTTP != nil && forHost.IngressRuleValue.HTTP.Paths != nil {
			for _, forPath := range forHost.IngressRuleValue.HTTP.Paths {
				l := l.
					With("kind", "rule").
					With("path", forPath.Path)

				path, err := ParsePath(forPath.Path, false)
				if err != nil {
					l.WithError(err).Warn("illegal path configured; ingress will not functioning; ignoring")
					continue
				}
				l = l.With("path", path)

				forService := forPath.Backend.Service
				if forService == nil {
					l.Warn("there is no service configured for path; ignoring")
					continue
				}
				if forService.Name == "" {
					l.Warn("there is no service.name configured for path; ignoring")
					continue
				}
				l = l.Withf("service", "%s/%s", source.namespace, forService.Name)

				if forService.Port.Number != 0 {
					l = l.With("port", forService.Port.Number)
				} else if forService.Port.Name != "" {
					l = l.With("port", forService.Port.Name)
				} else {
					l.Warn("there is neither a service.port.name nor service.port.number configured for path; ignoring")
					continue
				}

				pathType, err := ParsePathType(forPath.PathType)
				if err != nil {
					l.WithError(err).Warn("illegal pathType configured; ingress will not functioning; ignoring")
					continue
				}
				l = l.With("pathType", pathType)

				backend, err := instance.ingressToBackend(source, forPath.Backend, l)
				if err != nil {
					return err
				}
				if backend == nil {
					continue
				}

				options, err := instance.newOptionsBy(ingress)
				if err != nil {
					return err
				}

				host, err := instance.parseHost(&forHost, options)
				if err != nil {
					l.WithError(err).Warn("illegal host in ingress; ignoring")
					continue
				}

				r := NewRule(host, path, pathType, source, backend, options)
				if err := target.Put(r); err != nil {
					return err
				}
				l.Debug("element registered")
			}
		}
	}

	return nil
}

func (instance *repositoryImplState) parseHost(ingressRule *networkingv1.IngressRule, options Options) (value.WildcardSupportingFqdn, error) {
	plain := normalizeHostname(ingressRule.Host)
	var result value.WildcardSupportingFqdn
	if err := result.Set(plain); err != nil {
		return "", fmt.Errorf("illegal host %q: %w", plain, err)
	}
	return result, nil
}

func (instance *repositoryImplState) newOptionsBy(ingress *networkingv1.Ingress) (Options, error) {
	result := instance.OptionsFactory()
	if err := result.Set(ingress.GetAnnotations()); err != nil {
		return nil, err
	}

	return result, nil
}

func (instance *repositoryImplState) ingressToBackend(source *sourceReference, ib networkingv1.IngressBackend, usingLogger log.Logger) (net.Addr, error) {
	service, err := instance.ingressToService(source, ib)
	if err != nil {
		return nil, err
	}
	if service == nil {
		usingLogger.Warn("service not found; maybe orphan ingress?; ignoring")
		return nil, nil
	}

	if service.Spec.Type != v1.ServiceTypeClusterIP {
		usingLogger.
			With("serviceType", service.Spec.Type).
			Warn("unsupported serviceType; ignoring")
		return nil, nil
	}

	if strings.TrimSpace(service.Spec.ClusterIP) == "" {
		usingLogger.
			Warnf("serviceType is '%s' but clusterIP of service is not set; ignoring", v1.ServiceTypeClusterIP)
		return nil, nil
	}

	port, err := instance.evaluateServicePort(ib.Service.Port, service)
	if err != nil {
		usingLogger.
			WithError(err).
			Warn("cannot resolve backend port; ignoring")
		return nil, nil
	}

	addr, err := instance.clusterIpBasedServiceToAddr(service.Spec.ClusterIP, port)
	if err != nil {
		usingLogger.
			WithError(err).
			Warn("cannot resolve backend address; ignoring")
		return nil, nil
	}

	return addr, nil
}

func (instance *repositoryImplState) evaluateServicePort(in networkingv1.ServiceBackendPort, service *v1.Service) (int32, error) {
	if v := in.Name; v != "" {
		for _, candidate := range service.Spec.Ports {
			if candidate.Name == v {
				return candidate.Port, nil
			}
		}
		return 0, fmt.Errorf("unknown service reference %s:%s", service.Name, v)
	}
	return in.Number, nil
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

func (instance *repositoryImplState) ingressToService(source *sourceReference, ib networkingv1.IngressBackend) (*v1.Service, error) {
	serviceKey := fmt.Sprintf("%s/%s", source.namespace, ib.Service.Name)

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
