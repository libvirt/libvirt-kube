package api

import (
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	apiv1 "libvirt.org/libvirt-kube/pkg/api/v1alpha1"
)

type VirtnodeinfoClient struct {
	tpr TPRClient
}

func RegisterVirtnodeinfo(clientset *kubernetes.Clientset) error {
	err := RegisterResourceExtension(clientset, "virtnode.libvirt.org", "libvirt.org", "virtnodes", "v1alpha1", "Virt nodes")
	if err != nil {
		return err
	}

	RegisterResourceScheme("libvirt.org", "v1alpha1", &apiv1.Virtnode{}, &apiv1.VirtnodeList{})

	return nil
}

func NewVirtnodeinfoClient(namespace string, kubeconfig *rest.Config) (*VirtnodeinfoClient, error) {
	client, err := GetResourceClient(kubeconfig, "libvirt.org", "v1alpha1")
	if err != nil {
		return nil, err
	}
	return &VirtnodeinfoClient{
		tpr: TPRClient{
			ResourceName: "virtnodes",
			Namespace:    namespace,
			Rest:         client,
		},
	}, nil
}

func (c *VirtnodeinfoClient) List() (*apiv1.VirtnodeList, error) {
	var obj apiv1.VirtnodeList
	if err := c.tpr.Get("", &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

func (c *VirtnodeinfoClient) Watch() (watch.Interface, error) {
	return c.tpr.Watch()
}

func (c *VirtnodeinfoClient) Get(name string) (*apiv1.Virtnode, error) {
	var obj apiv1.Virtnode
	if err := c.tpr.Get(name, &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

func (c *VirtnodeinfoClient) Create(obj *apiv1.Virtnode) (*apiv1.Virtnode, error) {
	var newobj apiv1.Virtnode = *obj
	if err := c.tpr.Post(&newobj); err != nil {
		return nil, err
	}
	return &newobj, nil
}

func (c *VirtnodeinfoClient) Update(obj *apiv1.Virtnode) (*apiv1.Virtnode, error) {
	var newobj apiv1.Virtnode = *obj
	if err := c.tpr.Put(obj.Metadata.Name, &newobj); err != nil {
		return nil, err
	}
	return &newobj, nil
}

func (c *VirtnodeinfoClient) Delete(obj *apiv1.Virtnode) error {
	return c.tpr.Delete(obj.Metadata.Name)
}
