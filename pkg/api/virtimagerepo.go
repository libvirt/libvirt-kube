package api

import (
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"

	apiv1 "libvirt.org/libvirt-kube/pkg/api/v1alpha1"
)

type VirtimagerepoClient struct {
	tpr TPRClient
}

func RegisterVirtimagerepo(clientset *kubernetes.Clientset) error {
	err := RegisterResourceExtension(clientset, "virtimagerepo.libvirt.org", "libvirt.org", "virtimagerepos", "v1alpha1", "Virt image repositories")
	if err != nil {
		return err
	}

	RegisterResourceScheme("libvirt.org", "v1alpha1", &apiv1.Virtimagerepo{}, &apiv1.VirtimagerepoList{})

	return nil
}

func NewVirtimagerepoClient(namespace string, kubeconfig *rest.Config) (*VirtimagerepoClient, error) {
	client, err := GetResourceClient(kubeconfig, "libvirt.org", "v1alpha1")
	if err != nil {
		return nil, err
	}
	return &VirtimagerepoClient{
		tpr: TPRClient{
			ResourceName: "virtimagerepos",
			Namespace:    namespace,
			Rest:         client,
		},
	}, nil
}

func (c *VirtimagerepoClient) List() (*apiv1.VirtimagerepoList, error) {
	var obj apiv1.VirtimagerepoList
	if err := c.tpr.List(&obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

func (c *VirtimagerepoClient) Watch() (watch.Interface, error) {
	return c.tpr.Watch()
}

func (c *VirtimagerepoClient) Get(name string) (*apiv1.Virtimagerepo, error) {
	var obj apiv1.Virtimagerepo
	if err := c.tpr.Get(name, &obj); err != nil {
		return nil, err
	}
	return &obj, nil
}

func (c *VirtimagerepoClient) Create(obj *apiv1.Virtimagerepo) (*apiv1.Virtimagerepo, error) {
	var newobj apiv1.Virtimagerepo = *obj
	if err := c.tpr.Post(&newobj); err != nil {
		return nil, err
	}
	return &newobj, nil
}

func (c *VirtimagerepoClient) Update(obj *apiv1.Virtimagerepo) (*apiv1.Virtimagerepo, error) {
	var newobj apiv1.Virtimagerepo = *obj
	if err := c.tpr.Put(&newobj); err != nil {
		return nil, err
	}
	return &newobj, nil
}

func (c *VirtimagerepoClient) Delete(obj *apiv1.Virtimagerepo) error {
	return c.tpr.Delete(obj.Metadata.Name)
}
