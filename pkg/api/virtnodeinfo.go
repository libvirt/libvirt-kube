package api

import (
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

func (c *VirtnodeinfoClient) decode(res rest.Result) (*apiv1.Virtnode, error) {
	if err := res.Error(); err != nil {
		return nil, err
	}

	var obj apiv1.Virtnode
	res.Into(&obj)
	return &obj, nil

}

func (c *VirtnodeinfoClient) Get(name string) (*apiv1.Virtnode, error) {
	res := c.tpr.Get(name)
	return c.decode(res)
}

func (c *VirtnodeinfoClient) Create(obj *apiv1.Virtnode) (*apiv1.Virtnode, error) {
	res := c.tpr.Post(obj)
	return c.decode(res)
}

func (c *VirtnodeinfoClient) Update(obj *apiv1.Virtnode) (*apiv1.Virtnode, error) {
	res := c.tpr.Put(obj.Metadata.Name, obj)
	return c.decode(res)
}

func (c *VirtnodeinfoClient) Delete(obj *apiv1.Virtnode) error {
	res := c.tpr.Delete(obj.Metadata.Name)
	if err := res.Error(); err != nil {
		return err
	}
	return nil
}
