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

package nodeinfo

import (
	"time"

	"github.com/golang/glog"
	"github.com/libvirt/libvirt-go"
	"k8s.io/client-go/kubernetes"
	kubeapi "k8s.io/client-go/pkg/api"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"libvirt.org/libvirt-kube/pkg/api"
	apiv1 "libvirt.org/libvirt-kube/pkg/api/v1alpha1"
	"libvirt.org/libvirt-kube/pkg/hypervisor"
)

type Service struct {
	conn           *libvirt.Connect
	connNotify     chan hypervisor.ConnectEvent
	nodeinfo       *apiv1.Virtnode
	nodeinfoclient *api.VirtnodeinfoClient
}

func eventloop() {
	for {
		libvirt.EventRunDefaultImpl()
	}
}

func init() {
	libvirt.EventRegisterDefaultImpl()
	go eventloop()
}

func getKubeConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func NewService(libvirtURI string, kubeconfigfile string, nodename string) (*Service, error) {
	kubeconfig, err := getKubeConfig(kubeconfigfile)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	err = api.RegisterVirtnodeinfo(clientset)
	if err != nil {
		return nil, err
	}

	nodeinfoclient, err := api.NewVirtnodeinfoClient(kubeapi.NamespaceDefault, kubeconfig)
	if err != nil {
		return nil, err
	}

	nodeinfo, err := nodeinfoclient.Get(nodename)
	if err != nil {
		return nil, err
	}

	svc := &Service{
		connNotify:     make(chan hypervisor.ConnectEvent, 1),
		nodeinfoclient: nodeinfoclient,
		nodeinfo:       nodeinfo,
	}

	hypervisor.OpenConnect(libvirtURI, svc.connNotify)

	return svc, nil
}

func (s *Service) updateNode(phase apiv1.VirtnodePhase) error {
	s.nodeinfo.Status.Phase = phase
	if phase == apiv1.VirtnodeReady {
		err := VirtNodeUpdateFromHypervisor(s.nodeinfo, s.conn)
		if err != nil {
			s.nodeinfo.Status.Phase = apiv1.VirtnodeFailed
		}
	}

	glog.V(1).Info("Updating existing record")
	obj, err := s.nodeinfoclient.Update(s.nodeinfo)

	if err != nil {
		glog.Errorf("Unable to update node info %s", err)
		return err
	}
	glog.V(1).Infof("Result %s", obj)
	s.nodeinfo = obj

	return nil
}

func (s *Service) Run() error {
	glog.V(1).Info("Running node service")

	ticker := time.NewTicker(time.Second * 15)

	for {
		select {
		case hypEvent := <-s.connNotify:
			switch hypEvent.Type {
			case hypervisor.ConnectReady:
				glog.V(1).Info("Got connection ready event")
				s.conn = hypEvent.Conn
				s.updateNode(apiv1.VirtnodeReady)

			case hypervisor.ConnectFailed:
				s.conn.Close()
				s.conn = nil
				glog.V(1).Info("Got connection failed event")
				s.updateNode(apiv1.VirtnodeOffline)
			}
		case <-ticker.C:
			if s.conn != nil {
				glog.V(1).Info("Updating node info")
				s.updateNode(apiv1.VirtnodeReady)
			} else {
				glog.V(1).Info("Not connected, skipping update")
			}
		}

	}

	return nil
}
