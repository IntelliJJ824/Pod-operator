package client

import (
	//corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/rest"
	"github.com/pod-operator/tony-operator/pkg/apis/tony/v1alpha1"
)

type ServerInterface interface {
	Create(server *v1alpha1.Server) (*v1alpha1.Server, error)
}

type ServerClient struct {
	RestClient rest.Interface
	Ns string
}

func (client *ServerClient) Create(server *v1alpha1.Server) (*v1alpha1.Server, error) {
	result := v1alpha1.Server{}
	err := client.RestClient.Post().Namespace(client.Ns).Resource("Servers").Body(server).Do().Into(&result)
	return &result, err
}
