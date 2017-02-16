package api

import (
	"k8s.io/apimachinery/pkg/watch"
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

func (c *VirtimagefileClient) List() (*apiv1.VirtimagefileList, error) {
	var obj apiv1.VirtimagefileList
	if err := c.tpr.Get("", &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

func (c *VirtimagefileClient) Watch() (watch.Interface, error) {
	return c.tpr.Watch()
}

func (c *VirtimagefileClient) Get(name string) (*apiv1.Virtimagefile, error) {
	var obj apiv1.Virtimagefile
	if err := c.tpr.Get(name, &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

func (c *VirtimagefileClient) Create(obj *apiv1.Virtimagefile) (*apiv1.Virtimagefile, error) {
	var newobj apiv1.Virtimagefile = *obj
	if err := c.tpr.Post(&newobj); err != nil {
		return nil, err
	}
	return &newobj, nil
}

func (c *VirtimagefileClient) Update(obj *apiv1.Virtimagefile) (*apiv1.Virtimagefile, error) {
	var newobj apiv1.Virtimagefile = *obj
	if err := c.tpr.Put(obj.Metadata.Name, &newobj); err != nil {
		return nil, err
	}
	return &newobj, nil
}

func (c *VirtimagefileClient) Delete(obj *apiv1.Virtimagefile) error {
	return c.tpr.Delete(obj.Metadata.Name)
}
