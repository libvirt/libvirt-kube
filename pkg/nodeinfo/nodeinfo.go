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
	"github.com/libvirt/libvirt-go"
	"github.com/libvirt/libvirt-go-xml"
	kubeapi "libvirt.org/libvirt-kube/pkg/kubeapi/v1alpha1"
)

func VirtNodeFromHypervisor(conn *libvirt.Connect) (*kubeapi.VirtnodeSpec, error) {
	capsxml, err := conn.GetCapabilities()
	if err != nil {
		return nil, err
	}

	caps := libvirtxml.Caps{}
	if err = caps.Unmarshal(capsxml); err != nil {
		return nil, err
	}

	guests := make([]kubeapi.VirtnodeGuest, 0)

	for _, cguest := range caps.Guests {
		for _, cdom := range cguest.Arch.Domains {
			var cmachines []libvirtxml.CapsGuestMachine
			if cdom.Machines == nil {
				cmachines = cguest.Arch.Machines
			} else {
				cmachines = cdom.Machines
			}
			machines := make([]string, 0)
			for _, cmach := range cmachines {
				machines = append(machines, cmach.Name)
			}
			guests = append(guests, kubeapi.VirtnodeGuest{
				Hypervisor: cdom.Type,
				Arch:       cguest.Arch.Name,
				Type:       cguest.OSType,
				Machines:   machines,
			})
		}
	}

	cells := make([]kubeapi.VirtnodeNUMACell, 0)
	if caps.Host.NUMA != nil {
		for _, lvcell := range caps.Host.NUMA.Cells {
			ncpus := len(lvcell.CPUS)
			memory := make([]kubeapi.VirtnodeMemory, 0)
			if lvcell.PageInfo == nil {
				memory = append(memory, kubeapi.VirtnodeMemory{
					PageSize: 4096,
					Present:  lvcell.Memory.Size / 4,
				})
			} else {
				for _, lvpage := range lvcell.PageInfo {
					memory = append(memory, kubeapi.VirtnodeMemory{
						PageSize: lvpage.Size,
						Present:  lvpage.Count,
					})
				}
			}
			cells = append(cells, kubeapi.VirtnodeNUMACell{
				CPU: kubeapi.VirtnodeCPU{
					Avail: ncpus,
					Used:  0,
				},
				Memory: memory,
			})
		}
	} else {
		nodeinfo, err := conn.GetNodeInfo()
		if err != nil {
			return nil, err
		}
		ncpus := int(nodeinfo.Nodes * nodeinfo.Sockets * nodeinfo.Cores * nodeinfo.Threads)
		memory := make([]kubeapi.VirtnodeMemory, 0)
		memory = append(memory, kubeapi.VirtnodeMemory{
			PageSize: 4096,
			Present:  nodeinfo.Memory / 4,
		})
		cells = append(cells, kubeapi.VirtnodeNUMACell{
			CPU: kubeapi.VirtnodeCPU{
				Avail: ncpus,
				Used:  0,
			},
			Memory: memory,
		})
	}

	resources := kubeapi.VirtnodeResources{
		NUMACells: cells,
	}

	info := &kubeapi.VirtnodeSpec{
		UUID:      caps.Host.UUID,
		Arch:      caps.Host.CPU.Arch,
		Guests:    guests,
		Resources: resources,
	}

	return info, nil
}
