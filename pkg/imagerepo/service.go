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

package imagerepo

import (
	"time"

	"github.com/golang/glog"
	"github.com/libvirt/libvirt-go"
	"k8s.io/client-go/kubernetes"
	kubeapi "k8s.io/client-go/pkg/api"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"libvirt.org/libvirt-kube/pkg/api"
	"libvirt.org/libvirt-kube/pkg/hypervisor"
)

type Service struct {
	conn       *libvirt.Connect
	connNotify chan hypervisor.ConnectEvent
	clientset  *kubernetes.Clientset
	repo       *Repository
}

func getKubeConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func NewService(libvirtURI string, kubeconfigfile string, reponame string, repopath string) (*Service, error) {
	kubeconfig, err := getKubeConfig(kubeconfigfile)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	err = api.RegisterVirtimagerepo(clientset)
	if err != nil {
		return nil, err
	}

	err = api.RegisterVirtimagefile(clientset)
	if err != nil {
		return nil, err
	}

	imagerepoclient, err := api.NewVirtimagerepoClient(kubeapi.NamespaceDefault, kubeconfig)
	if err != nil {
		return nil, err
	}

	imagefileclient, err := api.NewVirtimagefileClient(kubeapi.NamespaceDefault, kubeconfig)
	if err != nil {
		return nil, err
	}

	imagerepo, err := imagerepoclient.Get(reponame)
	if err != nil {
		return nil, err
	}

	glog.V(1).Infof("Got repo %s", imagerepo)

	svc := &Service{
		connNotify: make(chan hypervisor.ConnectEvent, 1),
		clientset:  clientset,
		repo:       CreateRepository(imagerepoclient, imagefileclient, imagerepo, repopath),
	}

	hypervisor.OpenConnect(libvirtURI, svc.connNotify)

	return svc, nil
}

func (s *Service) Run() error {
	glog.V(1).Info("Running image repo service")

	ticker := time.NewTicker(time.Second * 15)

	err := s.repo.loadFileResources()
	if err != nil {
		return err
	}

	for {
		select {
		case hypEvent := <-s.connNotify:
			switch hypEvent.Type {
			case hypervisor.ConnectReady:
				glog.V(1).Info("Got connection ready event")
				s.conn = hypEvent.Conn
				s.repo.update(s.conn)

			case hypervisor.ConnectFailed:
				s.conn.Close()
				s.conn = nil
				glog.V(1).Info("Got connection failed event")
				s.repo.update(s.conn)
			}
		case <-ticker.C:
			if s.conn != nil {
				glog.V(1).Info("Updating repo")
				s.repo.update(s.conn)
			} else {
				glog.V(1).Info("Not connected, skipping update")
			}
		}
	}

	return nil
}
