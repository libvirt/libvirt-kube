package api

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	apiv1 "libvirt.org/libvirt-kube/pkg/api/v1alpha1"
)

type VirtimagefileClient struct {
	tpr TPRClient
}

func RegisterVirtimagefile(clientset *kubernetes.Clientset) error {
	err := RegisterResourceExtension(clientset, "virtimagefile.libvirt.org", "libvirt.org", "virtimagefiles", "v1alpha1", "Virt image files")
	if err != nil {
		return err
	}

	RegisterResourceScheme("libvirt.org", "v1alpha1", &apiv1.Virtimagefile{}, &apiv1.VirtimagefileList{})

	return nil
}

func NewVirtimagefileClient(namespace string, kubeconfig *rest.Config) (*VirtimagefileClient, error) {
	client, err := GetResourceClient(kubeconfig, "libvirt.org", "v1alpha1")
	if err != nil {
		return nil, err
	}
	return &VirtimagefileClient{
		tpr: TPRClient{
			ResourceName: "virtimagefiles",
			Namespace:    namespace,
			Rest:         client,
		},
	}, nil
}

func (c *VirtimagefileClient) decode(res rest.Result) (*apiv1.Virtimagefile, error) {
	if err := res.Error(); err != nil {
		return nil, err
	}

	var obj apiv1.Virtimagefile
	res.Into(&obj)
	return &obj, nil
}

func (c *VirtimagefileClient) List() (*apiv1.VirtimagefileList, error) {
	res := c.tpr.Get("")

	if err := res.Error(); err != nil {
		return nil, err
	}

	var obj apiv1.VirtimagefileList
	res.Into(&obj)
	return &obj, nil
}

func (c *VirtimagefileClient) Get(name string) (*apiv1.Virtimagefile, error) {
	res := c.tpr.Get(name)
	return c.decode(res)
}

func (c *VirtimagefileClient) Create(obj *apiv1.Virtimagefile) (*apiv1.Virtimagefile, error) {
	res := c.tpr.Post(obj)
	return c.decode(res)
}

func (c *VirtimagefileClient) Update(obj *apiv1.Virtimagefile) (*apiv1.Virtimagefile, error) {
	res := c.tpr.Put(obj.Metadata.Name, obj)
	return c.decode(res)
}

func (c *VirtimagefileClient) Delete(obj *apiv1.Virtimagefile) error {
	res := c.tpr.Delete(obj.Metadata.Name)
	if err := res.Error(); err != nil {
		return err
	}
	return nil
}
