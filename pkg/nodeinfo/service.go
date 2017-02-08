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
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	kubeapi "k8s.io/client-go/pkg/api"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"libvirt.org/libvirt-kube/pkg/api"
	apiv1 "libvirt.org/libvirt-kube/pkg/api/v1alpha1"
)

// Retry every 5 seconds for 30 seconds, then every 15 seconds
// for another minute, then every 60 seconds thereafter
var reconnectDelay = []int{
	5, 5, 5, 5, 5, 5, 15, 15, 15, 15, 60,
}

type Hypervisor struct {
	uri            string
	closed         chan libvirt.ConnectCloseReason
	conn           *libvirt.Connect
	reconnectDelay int
}

type Service struct {
	hypervisor     Hypervisor
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

func NewService(libvirtURI string, kubeconfigfile string) (*Service, error) {
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

	shim := &Service{
		hypervisor: Hypervisor{
			closed: make(chan libvirt.ConnectCloseReason, 1),
			uri:    libvirtURI,
		},
		nodeinfoclient: nodeinfoclient,
	}

	return shim, nil
}

func (s *Service) updateNode(phase apiv1.VirtnodePhase) error {
	if phase == apiv1.VirtnodeReady {
		nodeinfo, err := VirtNodeFromHypervisor(s.hypervisor.conn)
		if err != nil {
			if s.nodeinfo != nil {
				s.nodeinfo.Status.Phase = apiv1.VirtnodeFailed
			} else {
				glog.V(1).Info("No previous nodeinfo, returning")
				return nil
			}
		} else {
			s.nodeinfo = nodeinfo
		}
	} else {
		if s.nodeinfo != nil {
			s.nodeinfo.Status.Phase = phase
		} else {
			glog.V(1).Info("No previous nodeinfo, returning")
			return nil
		}
	}

	obj, err := s.nodeinfoclient.Get(s.nodeinfo.Metadata.Name)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if err != nil {
		glog.V(1).Info("Creating initial record")
		obj, err = s.nodeinfoclient.Create(s.nodeinfo)
	} else {
		glog.V(1).Info("Updating existing record")
		obj, err = s.nodeinfoclient.Update(s.nodeinfo)
	}

	if err != nil {
		glog.Errorf("Unable to update node info %s", err)
	} else {
		glog.V(1).Infof("Result %s", obj)
	}

	return nil
}

func (s *Service) libvirtClosed(conn *libvirt.Connect, reason libvirt.ConnectCloseReason) {
	glog.V(1).Infof("Notify about connection close %d", reason)
	s.hypervisor.closed <- reason
}

func (s *Service) connect() error {
	glog.V(1).Infof("Trying to connect to %s", s.hypervisor.uri)
	conn, err := libvirt.NewConnect(s.hypervisor.uri)
	if err != nil {
		return err
	}

	s.hypervisor.conn = conn
	s.hypervisor.closed = make(chan libvirt.ConnectCloseReason, 1)

	conn.RegisterCloseCallback(s.libvirtClosed)

	return nil
}

func (s *Service) disconnect() {
	glog.V(1).Info("Disconnecting from closed libvirt connection")
	s.hypervisor.conn.UnregisterCloseCallback()
	s.hypervisor.conn.Close()
	s.hypervisor.conn = nil
	s.hypervisor.reconnectDelay = 0
}

func (s *Service) Run() error {
	glog.V(1).Info("Running node service")
	for {
		select {
		case reason := <-s.hypervisor.closed:
			glog.V(1).Infof("Saw hypervisor disconnect reason %d", reason)
			s.disconnect()
			s.updateNode(apiv1.VirtnodeOffline)
		default:
			// Cause select to be non-blocking if not hv is not closed
		}

		if s.hypervisor.conn == nil {
			err := s.connect()
			if err != nil {
				glog.V(1).Infof("Unable to connect to %s, retry in %d seconds: %s",
					s.hypervisor.uri, reconnectDelay[s.hypervisor.reconnectDelay], err)
				time.Sleep(time.Duration(reconnectDelay[s.hypervisor.reconnectDelay]) * time.Second)
				if s.hypervisor.reconnectDelay < (len(reconnectDelay) - 1) {
					s.hypervisor.reconnectDelay++
				}
				s.hypervisor.reconnectDelay = 0
			}
		}

		if s.hypervisor.conn != nil {
			glog.V(1).Info("Updating node info")
			s.updateNode(apiv1.VirtnodeReady)
			time.Sleep(15 * time.Second)
		}
	}

	return nil
}
