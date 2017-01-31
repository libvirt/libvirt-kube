package api

import (
	"fmt"
	"net/http"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	kubeapi "k8s.io/client-go/pkg/api"
	"k8s.io/client-go/pkg/apis/extensions/v1beta1"
	"k8s.io/client-go/rest"
)

func RegisterResourceExtension(clientset *kubernetes.Clientset, uri string, group string, name string, version string, description string) error {
	tpr, err := clientset.Extensions().ThirdPartyResources().Get(uri, v1.GetOptions{})
	if err == nil {
		// Already exists !
		return nil
	}

	if !errors.IsNotFound(err) {
		return err
	}

	tpr = &v1beta1.ThirdPartyResource{
		ObjectMeta: v1.ObjectMeta{
			Name: uri,
		},
		Versions: []v1beta1.APIVersion{
			{Name: version},
		},
		Description: description,
	}

	_, err = clientset.Extensions().ThirdPartyResources().Create(tpr)
	if err != nil {
		return err
	}

	err = waitForResourceReady(clientset, group, name, version)

	return nil
}

func RegisterResourceScheme(group string, version string, obj, objlist runtime.Object) {
	schemeBuilder := runtime.NewSchemeBuilder(
		func(scheme *runtime.Scheme) error {
			scheme.AddKnownTypes(
				schema.GroupVersion{
					Group:   group,
					Version: version,
				},
				obj,
				objlist,
				&kubeapi.ListOptions{},
				&kubeapi.DeleteOptions{},
			)
			return nil
		})
	schemeBuilder.AddToScheme(kubeapi.Scheme)
}

func isResourceReady(clientset *kubernetes.Clientset, group string, name string, version string) (bool, error) {
	res := clientset.CoreV1().RESTClient().Get().AbsPath("apis", group, version, name).Do()
	err := res.Error()
	if err != nil {
		if se, ok := err.(*errors.StatusError); ok {
			if se.Status().Code == http.StatusNotFound {
				return false, nil
			}
		}
		return false, err
	}

	var statusCode int
	res.StatusCode(&statusCode)
	if statusCode != http.StatusOK {
		return false, fmt.Errorf("invalid status code: %d", statusCode)
	}

	return true, nil
}

func waitForResourceReady(clientset *kubernetes.Clientset, group string, name string, version string) error {
	return wait.Poll(3*time.Second, 30*time.Second, func() (bool, error) {
		return isResourceReady(clientset, group, name, version)
	})
}

func GetResourceClient(config *rest.Config, group string, version string) (*rest.RESTClient, error) {
	var tprconfig rest.Config
	tprconfig = *config

	tprconfig.GroupVersion = &schema.GroupVersion{
		Group:   group,
		Version: version,
	}
	tprconfig.APIPath = "/apis"
	tprconfig.ContentType = runtime.ContentTypeJSON
	tprconfig.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: kubeapi.Codecs}

	return rest.RESTClientFor(&tprconfig)
}
