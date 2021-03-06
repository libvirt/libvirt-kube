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
	"k8s.io/apimachinery/pkg/watch"
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

type TPRClient struct {
	ResourceName string
	Namespace    string
	Rest         *rest.RESTClient
}

type TPRObject interface {
	runtime.Object
	v1.ObjectMetaAccessor
}

type TPRObjectList interface {
	runtime.Object
	v1.ListMetaAccessor
}

func (c *TPRClient) List(obj TPRObjectList) error {
	res := c.Rest.Get().Resource(c.ResourceName).Namespace(c.Namespace).Do()
	if err := res.Error(); err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("Resource type '%s' name '%s' was not found", c.ResourceName, c.Namespace)
		}
		return err
	}
	return res.Into(obj)
}

func (c *TPRClient) Get(name string, obj TPRObject) error {
	var res rest.Result
	if name == "" {
		res = c.Rest.Get().Resource(c.ResourceName).Namespace(c.Namespace).Do()
	} else {
		res = c.Rest.Get().Resource(c.ResourceName).Namespace(c.Namespace).Name(name).Do()
	}
	if err := res.Error(); err != nil {
		if errors.IsNotFound(err) {
			return fmt.Errorf("Resource type '%s' name '%s/%s' was not found", c.ResourceName, c.Namespace, name)
		}
		return err
	}
	return res.Into(obj)
}

func (c *TPRClient) Put(obj TPRObject) error {
	name := obj.GetObjectMeta().GetName()
	res := c.Rest.Put().Resource(c.ResourceName).Namespace(c.Namespace).Name(name).Body(obj).Do()
	if err := res.Error(); err != nil {
		return err
	}
	return res.Into(obj)
}

func (c *TPRClient) Post(obj TPRObject) error {
	res := c.Rest.Post().Resource(c.ResourceName).Namespace(c.Namespace).Body(obj).Do()
	if err := res.Error(); err != nil {
		return err
	}
	return res.Into(obj)
}

func (c *TPRClient) Delete(name string) error {
	return c.Rest.Post().Resource(c.ResourceName).Namespace(c.Namespace).Name(name).Do().Error()
}

func (c *TPRClient) Watch() (watch.Interface, error) {
	return c.Rest.Get().Prefix("watch").Resource(c.ResourceName).Namespace(c.Namespace).Watch()
}
