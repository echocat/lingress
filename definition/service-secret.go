package definition

import (
	"fmt"
	"github.com/echocat/lingress/settings"
	"github.com/echocat/lingress/support"
	log "github.com/echocat/slf4g"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"strings"
	"time"
)

type ServiceSecret struct {
	*Definition

	settings    *settings.Settings
	client      kubernetes.Interface
	resyncAfter time.Duration
	namespace   string
}

func NewServiceSecrets(s *settings.Settings, client kubernetes.Interface, resyncAfter time.Duration, logger log.Logger) (*ServiceSecret, error) {
	if definition, err := newDefinition("service-secrets", nil, logger); err != nil {
		return nil, err
	} else {
		return &ServiceSecret{
			Definition:  definition,
			settings:    s,
			client:      client,
			resyncAfter: resyncAfter,
		}, nil
	}
}

func (this *ServiceSecret) SetNamespace(namespace string) {
	this.namespace = namespace
}

func (this *ServiceSecret) Init(stop support.Channel) error {
	if len(this.settings.Tls.SecretNames) == 0 &&
		this.settings.Tls.SecretNamePattern == nil &&
		len(this.settings.Tls.SecretLabelSelector) == 0 &&
		len(this.settings.Tls.SecretFieldSelector) == 0 {
		this.Logger.Info("Neither tls.secretNames nor tls.secretNamePatterns nor tls.secretLabelSelector nor tls.secretFieldSelector was specified. No service secret will be evaluated = No service specific TLS certificate will be available.")
		return nil
	}

	informerFactory := informers.NewSharedInformerFactoryWithOptions(
		this.client,
		this.resyncAfter,
		informers.WithTweakListOptions(func(options *metav1.ListOptions) {
			options.LabelSelector = strings.Join(this.settings.Tls.SecretLabelSelector, ",")
			options.FieldSelector = strings.Join(this.settings.Tls.SecretFieldSelector, ",")
		}),
	)
	this.SetInformer(informerFactory.Core().V1().Secrets().Informer())

	return this.Definition.Init(stop)
}

func (this *ServiceSecret) Get(key string) (*v1.Service, error) {
	if item, exists, err := this.informer.GetStore().GetByKey(key); err != nil {
		return nil, fmt.Errorf("cannot get secrets %s from cache: %v", key, err)
	} else if !exists {
		return nil, nil
	} else {
		return item.(*v1.Service), nil
	}
}
