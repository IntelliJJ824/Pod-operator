package clientset

import(
	"github.com/pod-operator/tony-operator/pkg/apis/tony/v1alpha1"
	"github.com/pod-operator/tony-operator/pkg/controller/client"
	"k8s.io/client-go/rest"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
)

type ServerV1Alpha1Interface interface {
	Servers(namespace string) client.ServerInterface
}

type ServerV1Alpha1Client struct {
	restClient rest.Interface
}

func NewForConfig(c *rest.Config) (*ServerV1Alpha1Client, error) {
	crdConfig := *c
	crdConfig.ContentConfig.GroupVersion = &schema.GroupVersion {
		Group: v1alpha1.GroupName,
		Version: v1alpha1.GroupVersion,
	}
	crdConfig.APIPath = "/apis"
	crdConfig.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}
	crdConfig.UserAgent = rest.DefaultKubernetesUserAgent()

	client, err := rest.RESTClientFor(&crdConfig)
	if err != nil {
		return nil, err
	}
	return &ServerV1Alpha1Client{restClient: client}, nil
}

func (c *ServerV1Alpha1Client) Servers(nameSpace string) client.ServerInterface {
	return &client.ServerClient {
		RestClient: c.restClient,
		Ns: nameSpace,
	}
}
