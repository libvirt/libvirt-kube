/*
 * This file is part of the libvirt-kube project
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Copyright 2017 Red Hat, Inc.
 *
 */

package kubenodeinfo

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/libvirt/libvirt-go"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	kubeapi "libvirt.org/libvirt-kube/pkg/kubeapi"
	kubeapiv1 "libvirt.org/libvirt-kube/pkg/kubeapi/v1alpha1"
	"libvirt.org/libvirt-kube/pkg/nodeinfo"
	"time"
)

type Service struct {
	hypervisor *libvirt.Connect
	clientset  *kubernetes.Clientset
	nodeinfo   *kubeapiv1.Virtnode
	tprclient  *rest.RESTClient
}

func getKubeConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func NewService(libvirtURI string, kubeconfigfile string) (*Service, error) {
	hypervisor, err := libvirt.NewConnect(libvirtURI)
	if err != nil {
		return nil, err
	}

	kubeconfig, err := getKubeConfig(kubeconfigfile)
	if err != nil {
		hypervisor.Close()
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		hypervisor.Close()
		return nil, err
	}

	err = kubeapi.RegisterResourceExtension(clientset, "virtnode.libvirt.org", "libvirt.org", "virtnodes", "v1alpha1", "Virt nodes")
	if err != nil {
		hypervisor.Close()
		return nil, err
	}

	kubeapi.RegisterResourceScheme("libvirt.org", "v1alpha1", &kubeapiv1.Virtnode{}, &kubeapiv1.VirtnodeList{})

	tprclient, err := kubeapi.GetResourceClient(kubeconfig, "libvirt.org", "v1alpha1")
	if err != nil {
		hypervisor.Close()
		return nil, err
	}

	shim := &Service{
		hypervisor: hypervisor,
		clientset:  clientset,
		tprclient:  tprclient,
	}

	return shim, nil
}

func (s *Service) updateNode() error {

	nodeinfo, err := nodeinfo.VirtNodeFromHypervisor(s.hypervisor)
	if err != nil {
		return err
	}
	fmt.Println(nodeinfo)

	res := s.tprclient.Get().Resource("virtnodes").Namespace(api.NamespaceDefault).Name(nodeinfo.Metadata.Name).Do()
	err = res.Error()

	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		res = s.tprclient.Post().Resource("virtnodes").Namespace(api.NamespaceDefault).Body(nodeinfo).Do()
	} else {
		res = s.tprclient.Put().Resource("virtnodes").Namespace(api.NamespaceDefault).Name(nodeinfo.Metadata.Name).Body(nodeinfo).Do()
	}

	err = res.Error()
	if err != nil {
		glog.Errorf("Unable to update node info %s", err)
	} else {
		var result kubeapiv1.Virtnode
		res.Into(&result)
		glog.V(1).Infof("Result %s", result)
	}

	return nil
}

func (s *Service) Run() error {
	for {
		glog.V(1).Info("Updating node info")
		s.updateNode()
		time.Sleep(15 * time.Second)
	}

	return nil
}
