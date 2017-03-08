package api

import (
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	apiv1 "libvirt.org/libvirt-kube/pkg/api/v1alpha1"
)

type VirtmachineClient struct {
	tpr TPRClient
}

func RegisterVirtmachine(clientset *kubernetes.Clientset) error {
	err := RegisterResourceExtension(clientset, "virtmachine.libvirt.org", "libvirt.org", "virtmachines", "v1alpha1", "Virt machines")
	if err != nil {
		return err
	}

	RegisterResourceScheme("libvirt.org", "v1alpha1", &apiv1.Virtmachine{}, &apiv1.VirtmachineList{})

	return nil
}

func NewVirtmachineClient(namespace string, kubeconfig *rest.Config) (*VirtmachineClient, error) {
	client, err := GetResourceClient(kubeconfig, "libvirt.org", "v1alpha1")
	if err != nil {
		return nil, err
	}
	return &VirtmachineClient{
		tpr: TPRClient{
			ResourceName: "virtmachines",
			Namespace:    namespace,
			Rest:         client,
		},
	}, nil
}

func (c *VirtmachineClient) List() (*apiv1.VirtmachineList, error) {
	var obj apiv1.VirtmachineList
	if err := c.tpr.List(&obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

func (c *VirtmachineClient) Watch() (watch.Interface, error) {
	return c.tpr.Watch()
}

func (c *VirtmachineClient) Get(name string) (*apiv1.Virtmachine, error) {
	var obj apiv1.Virtmachine
	if err := c.tpr.Get(name, &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

func (c *VirtmachineClient) Create(obj *apiv1.Virtmachine) (*apiv1.Virtmachine, error) {
	var newobj apiv1.Virtmachine = *obj
	if err := c.tpr.Post(&newobj); err != nil {
		return nil, err
	}
	return &newobj, nil
}

func (c *VirtmachineClient) Update(obj *apiv1.Virtmachine) (*apiv1.Virtmachine, error) {
	var newobj apiv1.Virtmachine = *obj
	if err := c.tpr.Put(&newobj); err != nil {
		return nil, err
	}
	return &newobj, nil
}

func (c *VirtmachineClient) Delete(obj *apiv1.Virtmachine) error {
	return c.tpr.Delete(obj.Metadata.Name)
}
