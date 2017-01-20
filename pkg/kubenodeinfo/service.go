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
	kubeapi "libvirt.org/libvirt-kube/pkg/kubeapi/v1alpha1"
	"libvirt.org/libvirt-kube/pkg/nodeinfo"
	"time"
)

type Service struct {
	hypervisor *libvirt.Connect
	nodeinfo   *kubeapi.Virtnode
}

func NewService(libvirtURI string) (*Service, error) {
	hypervisor, err := libvirt.NewConnect(libvirtURI)
	if err != nil {
		return nil, err
	}

	shim := &Service{
		hypervisor: hypervisor,
	}

	return shim, nil
}

func (s *Service) updateNodeInfo() error {

	node, err := nodeinfo.VirtNodeFromHypervisor(s.hypervisor)
	if err != nil {
		return err
	}

	fmt.Println(node)

	return nil
}

func (s *Service) Run() error {
	for {
		glog.V(1).Info("Updating node info")
		s.updateNodeInfo()
		time.Sleep(60 * time.Second)
	}

	return nil
}
