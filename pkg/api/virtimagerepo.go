package api

import (
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

func (c *VirtimagerepoClient) decode(res rest.Result) (*apiv1.Virtimagerepo, error) {
	if err := res.Error(); err != nil {
		return nil, err
	}

	var obj apiv1.Virtimagerepo
	res.Into(&obj)
	return &obj, nil

}

func (c *VirtimagerepoClient) Get(name string) (*apiv1.Virtimagerepo, error) {
	res := c.tpr.Get(name)
	return c.decode(res)
}

func (c *VirtimagerepoClient) Create(obj *apiv1.Virtimagerepo) (*apiv1.Virtimagerepo, error) {
	res := c.tpr.Post(obj)
	return c.decode(res)
}

func (c *VirtimagerepoClient) Update(obj *apiv1.Virtimagerepo) (*apiv1.Virtimagerepo, error) {
	res := c.tpr.Put(obj.Metadata.Name, obj)
	return c.decode(res)
}

func (c *VirtimagerepoClient) Delete(obj *apiv1.Virtimagerepo) error {
	res := c.tpr.Delete(obj.Metadata.Name)
	if err := res.Error(); err != nil {
		return err
	}
	return nil
}
