package rules

import (
	"fmt"
	"github.com/echocat/lingress/definition"
	"github.com/echocat/lingress/kubernetes"
	"github.com/echocat/lingress/settings"
	"github.com/echocat/lingress/support"
	"github.com/echocat/lingress/value"
	"github.com/echocat/slf4g"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net"
	"strings"
	"sync/atomic"
)

type Query struct {
	Host value.Fqdn
	Path string
}

type CertificateQuery struct {
	Host value.WildcardSupportingFqdn
}

type Repository interface {
	Init(stop support.Channel) error
	All(consumer func(Rule) error) error
	FindBy(Query) (Rules, error)
}

type CombinedRepository interface {
	Repository
	CertificateRepository
}

type KubernetesBasedRepository struct {
	settings *settings.Settings

	Environment *kubernetes.Environment
	ByHostRules *ByHost
	Logger      log.Logger

	CertificatesByHost CertificatesByHost
	OptionsFactory     OptionsFactory
}

func NewRepository(s *settings.Settings, logger log.Logger) (CombinedRepository, error) {
	environment, err := kubernetes.NewEnvironment(s)
	if err != nil {
		return nil, err
	}
	result := &KubernetesBasedRepository{
		settings:    s,
		Environment: environment,

		OptionsFactory:     DefaultOptionsFactory,
		CertificatesByHost: CertificatesByHost{},

		Logger: logger,
	}
	result.ByHostRules = NewByHost(result.onRuleAdded, result.onRuleRemoved)
	return result, nil
}

func (this *KubernetesBasedRepository) Init(stop support.Channel) error {
	log.Info("Initial sync of definitions...")

	client, err := this.Environment.NewClient()
	if err != nil {
		return err
	}
	definitions, err := definition.New(this.settings, client, this.settings.Discovery.ResyncAfter, this.Logger)
	if err != nil {
		return err
	}
	definitions.SetNamespace(this.settings.Kubernetes.Namespace)

	state := &repositoryImplState{
		KubernetesBasedRepository: this,
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

	log.Info("Initial sync of definitions... done!")
	return nil
}

func (this *KubernetesBasedRepository) onRuleAdded(_ []string, r Rule) {
	this.Logger.With("rule", r).Debug("Rule added.")
}

func (this *KubernetesBasedRepository) onRuleRemoved(_ []string, r Rule) {
	this.Logger.With("rule", r).Debug("Rule removed.")
}

type repositoryImplState struct {
	*KubernetesBasedRepository

	definitions *definition.Definitions
	initiated   atomic.Value
}

func (this *repositoryImplState) onSecretCertificatesChanged(ref support.ObjectReference, new metav1.Object) error {
	l := this.Logger.
		With("ref", ref)

	if removed, err := this.CertificatesByHost.RemoveBySource(ref); err != nil {
		return err
	} else if len(removed) > 0 {
		l.With("fqdns", removed).Info("Certificates for FQNDs removed.")
	}

	if new == nil {
		return nil
	}

	s := new.(*v1.Secret)

	for file, candidate := range s.Data {
		l := l.With("certificate", file)
		base, ext := support.SplitExt(file)
		if ext == ".crt" || ext == ".cer" {
			privateKeyFile := base + ".key"
			l := l.With("privateKey", privateKeyFile)
			pk, ok := s.Data[privateKeyFile]
			if !ok {
				l.Debug("Cannot find expected privateKey in secret for provided certificate; ignoring...")
				continue
			}
			if base == "tls" {
				ca, ok := s.Data["ca.cert"]
				if ok {
					candidate = append(ca, '\n')
					candidate = append(ca, ca...)
				}
			}
			if added, err := this.CertificatesByHost.AddBytes(ref, candidate, pk); err != nil {
				l.WithError(err).Warn("Cannot parse certificate and privateKey pair from secret; ignoring...")
				continue
			} else if len(added) > 0 {
				l.With("fqdns", added).Info("Certificates for FQNDs added.")
			}
		}
	}

	return nil
}

func (this *repositoryImplState) isExpectedCertificatesKey(what support.ObjectReference) bool {
	if this.settings.Tls.SecretNamePattern != nil && !this.settings.Tls.SecretNamePattern.MatchString(what.ShortString()) {
		return false
	}

	if len(this.settings.Tls.SecretNames) > 0 {
		for _, candidate := range this.settings.Tls.SecretNames {
			if !strings.Contains(candidate, "/") {
				candidate = this.settings.Kubernetes.Namespace + "/" + candidate
			}
			if candidate == what.ShortString() {
				return true
			}
		}
		return false
	}

	return true
}

func (this *repositoryImplState) onServiceSecretsElementAdded(ref support.ObjectReference, new metav1.Object) error {
	if this.isExpectedCertificatesKey(ref) {
		return this.onSecretCertificatesChanged(ref, new)
	}
	return nil
}

func (this *repositoryImplState) onServiceSecretsElementUpdated(ref support.ObjectReference, _, new metav1.Object) error {
	if this.isExpectedCertificatesKey(ref) {
		return this.onSecretCertificatesChanged(ref, new)
	}
	return nil
}

func (this *repositoryImplState) onServiceSecretsElementRemoved(ref support.ObjectReference) error {
	if this.isExpectedCertificatesKey(ref) {
		return this.onSecretCertificatesChanged(ref, nil)
	}
	return nil
}

func (this *repositoryImplState) onIngressElementAdded(ref support.ObjectReference, new metav1.Object) error {
	target := this.ByHostRules
	clonedUpdate := this.initiated.Load() == true
	if clonedUpdate {
		target = target.Clone()
	}

	candidate := new.(*networkingv1.Ingress)

	if !this.matchesIngressClass(candidate) {
		return nil
	}

	if err := this.visitIngress(ref, candidate, target); err != nil {
		return err
	}

	if clonedUpdate {
		this.ByHostRules = target
	}
	return nil
}

func (this *repositoryImplState) onIngressElementUpdated(ref support.ObjectReference, _, new metav1.Object) error {
	candidate := new.(*networkingv1.Ingress)

	if this.matchesIngressClass(candidate) {
		return this.onIngressElementAdded(ref, new)
	}

	return this.onIngressElementRemoved(ref)
}

func (this *repositoryImplState) onIngressElementRemoved(ref support.ObjectReference) error {
	target := this.ByHostRules
	clonedUpdate := this.initiated.Load() != true
	if clonedUpdate {
		target = target.Clone()
	}

	if err := target.Remove(PredicateByObjectReference(ref)); err != nil {
		return fmt.Errorf("cannot remove previous element by source %v: %v", ref, err)
	}

	if clonedUpdate {
		this.ByHostRules = target
	}
	return nil
}

func (this *repositoryImplState) matchesIngressClass(what *networkingv1.Ingress) bool {
	requested := what.Spec.IngressClassName
	classes := this.settings.Ingress.GetClasses()
	if requested == nil {
		for _, candidate := range classes {
			if candidate == "" || candidate == "*" {
				return true
			}
		}
		return false
	}
	for _, candidate := range classes {
		if candidate == *requested || candidate == "*" {
			return true
		}
	}
	return false
}

func (this *repositoryImplState) visitIngress(ref support.ObjectReference, ingress *networkingv1.Ingress, target *ByHost) error {
	if err := target.Remove(PredicateByObjectReference(ref)); err != nil {
		return err
	}

	l := this.Logger.
		With("ref", ref)

	if len(ingress.Spec.TLS) > 0 {
		l.Warn("Currently ingress configurations with spec.tls settings are not supported; ignoring...")

		return nil
	}

	if v := ingress.Spec.DefaultBackend; v != nil {
		l := l.With("kind", "defaultBackend")
		backend, err := this.ingressToBackend(ref, v, l)
		if err != nil {
			return err
		}
		if backend != nil {
			options, err := this.newOptionsBy(ingress)
			if err != nil {
				return err
			}
			r := NewRule("", []string{}, PathTypePrefix, ref, backend, options)
			if err := target.Put(r); err != nil {
				return err
			}
			l.Debug("Element registered.")
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
					l.WithError(err).Warn("Illegal path configured; ingress will not functioning; ignoring...")
					continue
				}
				l = l.With("path", path)

				if forPath.Backend.Resource != nil {
					l.Warn("Currently ingress configurations with spec.rules.http.paths.backend.resource settings are not supported; ignoring...")
					continue
				}

				forService := forPath.Backend.Service
				if forService == nil {
					l.Warn("There is no service configured for path; ignoring...")
					continue
				}
				if forService.Name == "" {
					l.Warn("There is no service.name configured for path; ignoring...")
					continue
				}
				serviceOr := this.ingressToServiceReference(ref, forService)
				l = l.With("service", serviceOr)

				if forService.Port.Number != 0 {
					l = l.With("port", forService.Port.Number)
				} else if forService.Port.Name != "" {
					l = l.With("port", forService.Port.Name)
				} else {
					l.Warn("There is neither a service.port.name nor service.port.number configured for path; ignoring...")
					continue
				}

				pathType, err := ParsePathType(forPath.PathType)
				if err != nil {
					l.WithError(err).Warn("Illegal pathType configured; ingress will not functioning; ignoring...")
					continue
				}
				l = l.With("pathType", pathType)

				backend, err := this.ingressToBackend(ref, &forPath.Backend, l)
				if err != nil {
					return err
				}
				if backend == nil {
					continue
				}

				options, err := this.newOptionsBy(ingress)
				if err != nil {
					return err
				}

				host, err := this.parseHost(&forHost, options)
				if err != nil {
					l.WithError(err).Warn("Illegal host in ingress; ignoring...")
					continue
				}

				r := NewRule(host, path, pathType, ref, backend, options)
				if err := target.Put(r); err != nil {
					return err
				}
				l.Debug("Element registered.")
			}
		}
	}

	return nil
}

func (this *repositoryImplState) parseHost(ingressRule *networkingv1.IngressRule, _ Options) (value.WildcardSupportingFqdn, error) {
	plain := normalizeHostname(ingressRule.Host)
	var result value.WildcardSupportingFqdn
	if err := result.Set(plain); err != nil {
		return "", fmt.Errorf("illegal host %q: %w", plain, err)
	}
	return result, nil
}

func (this *repositoryImplState) newOptionsBy(ingress *networkingv1.Ingress) (Options, error) {
	result := this.OptionsFactory()
	if err := result.Set(ingress.GetAnnotations()); err != nil {
		return nil, err
	}

	return result, nil
}

func (this *repositoryImplState) ingressToBackend(source support.ObjectReference, ib *networkingv1.IngressBackend, usingLogger log.Logger) (net.Addr, error) {
	service, err := this.ingressToService(source, ib)
	if err != nil {
		return nil, err
	}
	if service == nil {
		usingLogger.Warn("Service not found; maybe orphan ingress?; ignoring...")
		return nil, nil
	}

	if service.Spec.Type != v1.ServiceTypeClusterIP && service.Spec.Type != "" {
		usingLogger.
			With("serviceType", service.Spec.Type).
			Warn("Unsupported serviceType; ignoring...")
		return nil, nil
	}

	if strings.TrimSpace(service.Spec.ClusterIP) == "" {
		usingLogger.
			Warnf("serviceType is '%s' but clusterIP of service is not set; ignoring.", v1.ServiceTypeClusterIP)
		return nil, nil
	}

	port, err := this.evaluateServicePort(ib.Service.Port, service)
	if err != nil {
		usingLogger.
			WithError(err).
			Warn("Cannot resolve backend port; ignoring...")
		return nil, nil
	}

	addr, err := this.clusterIpBasedServiceToAddr(service.Spec.ClusterIP, port)
	if err != nil {
		usingLogger.
			WithError(err).
			Warn("Cannot resolve backend address; ignoring...")
		return nil, nil
	}

	return addr, nil
}

func (this *repositoryImplState) evaluateServicePort(in networkingv1.ServiceBackendPort, service *v1.Service) (int32, error) {
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

func (this *repositoryImplState) clusterIpBasedServiceToAddr(ipStr string, port int32) (net.Addr, error) {
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

func (this *repositoryImplState) ingressToServiceReference(source support.ObjectReference, isb *networkingv1.IngressServiceBackend) support.ObjectReference {
	return source.WithApiVersionAndKind("v1", "Service").WithName(isb.Name)
}

func (this *repositoryImplState) ingressToService(source support.ObjectReference, ib *networkingv1.IngressBackend) (*v1.Service, error) {
	return this.definitions.Service.Get(this.ingressToServiceReference(source, ib.Service).ShortString())
}

func (this *KubernetesBasedRepository) All(consumer func(Rule) error) error {
	return this.ByHostRules.All(consumer)
}

func (this *KubernetesBasedRepository) FindBy(q Query) (Rules, error) {
	host := q.Host
	path, err := ParsePath(q.Path, true)
	if err != nil {
		return nil, err
	}
	return this.ByHostRules.Find(host, path)
}

func (this *KubernetesBasedRepository) FindCertificatesBy(q CertificateQuery) (Certificates, error) {
	return this.CertificatesByHost.Find(q.Host), nil
}
