package kubernetes

import (
	"github.com/echocat/lingress/support"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"testing"
)

func Test_NewEnvironment_succeeds(t *testing.T) {
	g := NewGomegaWithT(t)

	defer unsetEnvVarTemporary(EnvVarKubeconfig)()
	instance, err := NewEnvironment()

	g.Expect(err).To(BeNil())
	g.Expect(instance).NotTo(BeNil())

	g.Expect(instance.Kubeconfig.IsEmpty()).To(BeTrue())
	g.Expect(instance.Context).To(Equal(""))
	g.Expect(instance.payload).To(BeNil())
}

func Test_Environment_get_with_default_values_succeeds(t *testing.T) {
	g := NewGomegaWithT(t)

	defer unsetEnvVarTemporary(EnvVarKubeconfig)()
	env := MustNewEnvironment()
	env.Kubeconfig.ResolveDefaultLocation = func() string { return "resources/kubeconfig_two_contexts.yml" }

	instance, err := env.get()

	g.Expect(err).To(BeNil())
	g.Expect(instance).NotTo(BeNil())
	g.Expect(instance.restConfig).ToNot(BeNil())
	g.Expect(instance.restConfig.Host).To(Equal("http://127.0.0.1:8080"))
	g.Expect(instance.contextName).To(Equal("context1"))
}

func Test_Environment_get_with_explicit_context_succeeds(t *testing.T) {
	g := NewGomegaWithT(t)

	defer unsetEnvVarTemporary(EnvVarKubeconfig)()
	env := MustNewEnvironment()
	env.Kubeconfig.ResolveDefaultLocation = func() string { return "resources/kubeconfig_two_contexts.yml" }
	env.Context = "context2"

	instance, err := env.get()

	g.Expect(err).To(BeNil())
	g.Expect(instance).NotTo(BeNil())
	g.Expect(instance.restConfig).ToNot(BeNil())
	g.Expect(instance.restConfig.Host).To(Equal("http://127.0.0.2:8080"))
	g.Expect(instance.contextName).To(Equal("context2"))
}

//noinspection GoUnhandledErrorResult
func Test_Environment_get_from_envVar_succeeds(t *testing.T) {
	g := NewGomegaWithT(t)

	defer setEnvVarTemporaryToFileContent(EnvVarKubeconfig, "resources/kubeconfig_alternative.yml")()
	env := MustNewEnvironment()
	env.Kubeconfig.ResolveDefaultLocation = func() string { return "resources/kubeconfig_two_contexts.yml" }

	instance, err := env.get()

	g.Expect(err).To(BeNil())
	g.Expect(instance).NotTo(BeNil())
	g.Expect(instance.restConfig).ToNot(BeNil())
	g.Expect(instance.restConfig.Host).To(Equal("http://127.0.0.3:8080"))
	g.Expect(instance.contextName).To(Equal("context3"))
}

func Test_Environment_get_without_current_context_fails(t *testing.T) {
	g := NewGomegaWithT(t)

	defer unsetEnvVarTemporary(EnvVarKubeconfig)()
	env := MustNewEnvironment()
	env.Kubeconfig.ResolveDefaultLocation = func() string { return "resources/kubeconfig_without_current_context.yml" }
	env.Context = ""

	_, err := env.get()

	g.Expect(err).To(Equal(clientcmd.ErrNoContext))
}

func Test_Environment_get_with_nonExisting_context_fails(t *testing.T) {
	g := NewGomegaWithT(t)

	defer unsetEnvVarTemporary(EnvVarKubeconfig)()
	env := MustNewEnvironment()
	env.Kubeconfig.ResolveDefaultLocation = func() string { return "resources/kubeconfig_two_contexts.yml" }
	env.Context = "foobar"

	_, err := env.get()

	g.Expect(err).To(MatchError(`context "foobar" does not exist`))
}

func Test_Environment_get_with_nonExisting_kubeconfigFile_fails(t *testing.T) {
	g := NewGomegaWithT(t)

	defer unsetEnvVarTemporary(EnvVarKubeconfig)()
	env := MustNewEnvironment()
	env.Kubeconfig.value = "nonExisting.yml"

	_, err := env.get()

	g.Expect(os.IsNotExist(err)).To(BeTrue(), "Should IsNotExist error - but got %v", err)
}

func Test_Environment_get_with_nonExisting_defaultConfig_and_envVar_fails(t *testing.T) {
	g := NewGomegaWithT(t)

	defer unsetEnvVarTemporary(EnvVarKubeconfig)()
	env := MustNewEnvironment()
	env.Kubeconfig.ResolveDefaultLocation = func() string { return "nonExisting.yml" }

	_, err := env.get()

	g.Expect(err).To(MatchError("there is neither argument --kubeconfig nor environment variable KUBE_CONFIG provided nor does nonExisting.yml exist"))
}

func Test_Environment_get_with_mock_with_set_context_succeeds(t *testing.T) {
	g := NewGomegaWithT(t)

	defer unsetEnvVarTemporary(EnvVarKubeconfig)()
	env := MustNewEnvironment()
	env.Context = "foobar"
	err := env.Kubeconfig.Set(KubeconfigMock)
	g.Expect(err).To(BeNil())

	instance, err := env.get()

	g.Expect(err).To(BeNil())
	g.Expect(instance).ToNot(BeNil())
	g.Expect(instance.restConfig).To(BeNil())
	g.Expect(instance.contextName).To(Equal("foobar"))
}

func Test_Environment_get_with_mock_with_empty_context_succeeds(t *testing.T) {
	g := NewGomegaWithT(t)

	defer unsetEnvVarTemporary(EnvVarKubeconfig)()
	env := MustNewEnvironment()
	err := env.Kubeconfig.Set(KubeconfigMock)
	g.Expect(err).To(BeNil())

	instance, err := env.get()

	g.Expect(err).To(BeNil())
	g.Expect(instance).ToNot(BeNil())
	g.Expect(instance.restConfig).To(BeNil())
	g.Expect(instance.contextName).To(Equal("mock"))
}

func Test_Environment_get_with_incluster_succeeds(t *testing.T) {
	g := NewGomegaWithT(t)

	defer unsetEnvVarTemporary(EnvVarKubeconfig)()
	defer setEnvVarTemporaryTo("KUBERNETES_SERVICE_HOST", "127.0.0.66")()
	defer setEnvVarTemporaryTo("KUBERNETES_SERVICE_PORT", "8080")()
	env := MustNewEnvironment()
	env.tokenFile = "resources/serviceaccount_token"
	env.rootCAFile = "resources/serviceaccount_ca.crt"
	env.namespaceFile = "resources/serviceaccount_namespace"
	err := env.Kubeconfig.Set(KubeconfigInCluster)
	g.Expect(err).To(BeNil())

	instance, err := env.get()

	g.Expect(err).To(BeNil())
	g.Expect(instance).ToNot(BeNil())
	g.Expect(instance.restConfig).NotTo(BeNil())
	g.Expect(instance.contextName).To(Equal("incluster"))
	g.Expect(env.Namespace).To(Equal("aNamespace"))
}

func Test_Environment_get_with_incluster_without_envVarServicePort_fails(t *testing.T) {
	g := NewGomegaWithT(t)

	defer unsetEnvVarTemporary(EnvVarKubeconfig)()
	defer setEnvVarTemporaryTo("KUBERNETES_SERVICE_HOST", "127.0.0.66")()
	defer unsetEnvVarTemporary("KUBERNETES_SERVICE_PORT")()
	env := MustNewEnvironment()
	env.tokenFile = "resources/serviceaccount_token"
	env.rootCAFile = "resources/serviceaccount_ca.crt"
	err := env.Kubeconfig.Set(KubeconfigInCluster)
	g.Expect(err).To(BeNil())

	_, err = env.get()

	g.Expect(err).To(MatchError("unable to load in-cluster configuration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined"))
}

func Test_Environment_get_with_incluster_without_envVarServiceHost_fails(t *testing.T) {
	g := NewGomegaWithT(t)

	defer unsetEnvVarTemporary(EnvVarKubeconfig)()
	defer unsetEnvVarTemporary("KUBERNETES_SERVICE_HOST")()
	defer setEnvVarTemporaryTo("KUBERNETES_SERVICE_PORT", "8080")()
	env := MustNewEnvironment()
	env.tokenFile = "resources/serviceaccount_token"
	env.rootCAFile = "resources/serviceaccount_ca.crt"
	err := env.Kubeconfig.Set(KubeconfigInCluster)
	g.Expect(err).To(BeNil())

	_, err = env.get()

	g.Expect(err).To(MatchError("unable to load in-cluster configuration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined"))
}

func Test_Environment_get_with_incluster_without_token_fails(t *testing.T) {
	g := NewGomegaWithT(t)

	defer unsetEnvVarTemporary(EnvVarKubeconfig)()
	defer setEnvVarTemporaryTo("KUBERNETES_SERVICE_HOST", "127.0.0.66")()
	defer setEnvVarTemporaryTo("KUBERNETES_SERVICE_PORT", "8080")()
	env := MustNewEnvironment()
	env.tokenFile = "nonExisting"
	env.rootCAFile = "resources/serviceaccount_ca.crt"
	err := env.Kubeconfig.Set(KubeconfigInCluster)
	g.Expect(err).To(BeNil())

	_, err = env.get()

	g.Expect(err).To(MatchError(HavePrefix(`failed to read token file "nonExisting": open nonExisting:`)))
}

func Test_Environment_get_with_incluster_without_rootCa_fails(t *testing.T) {
	g := NewGomegaWithT(t)

	defer unsetEnvVarTemporary(EnvVarKubeconfig)()
	defer setEnvVarTemporaryTo("KUBERNETES_SERVICE_HOST", "127.0.0.66")()
	defer setEnvVarTemporaryTo("KUBERNETES_SERVICE_PORT", "8080")()
	env := MustNewEnvironment()
	env.tokenFile = "resources/serviceaccount_token"
	env.rootCAFile = "nonExisting"
	err := env.Kubeconfig.Set(KubeconfigInCluster)
	g.Expect(err).To(BeNil())

	_, err = env.get()

	g.Expect(err).To(MatchError(HavePrefix(`expected to load root CA config from nonExisting, but got err:`)))
}

func setEnvVarTemporaryTo(key, value string) (rollback envVarRollback) {
	if oldValue, oldContentExists := os.LookupEnv(key); oldContentExists {
		rollback = func() {
			_ = os.Setenv(key, oldValue)
		}
	} else {
		rollback = func() {
			_ = os.Unsetenv(key)
		}
	}
	_ = os.Setenv(key, value)
	return
}

func unsetEnvVarTemporary(key string) (rollback envVarRollback) {
	if oldValue, oldContentExists := os.LookupEnv(key); oldContentExists {
		rollback = func() {
			_ = os.Setenv(key, oldValue)
		}
	} else {
		rollback = func() {
			_ = os.Unsetenv(key)
		}
	}
	_ = os.Unsetenv(key)
	return
}

func setEnvVarTemporaryToFileContent(key, filename string) (rollback envVarRollback) {
	value, err := ioutil.ReadFile(filename)
	support.Must(err)
	return setEnvVarTemporaryTo(key, string(value))
}

type envVarRollback func()
