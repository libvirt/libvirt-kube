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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
)

// Virttemplate defines a Virttemplate deployment.
type Virttemplate struct {
	metav1.TypeMeta `json:",inline"`
	v1.ObjectMeta   `json:"metadata,omitempty"`
	Spec            VirttemplateSpec `json:"spec"`
	//	Status          *VirttemplateStatus `json:"status,omitempty"`
}

// VirttemplateList is a list of Virttemplates.
type VirttemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []*Virttemplate `json:"items"`
}

// VirttemplateSpec holds specification parameters of a Virttemplate deployment.
type VirttemplateSpec struct {
	// Hypervisor type (libvirt: /domain/@type)
	Type    string             `json:"type"`
	Arch    string             `json:"arch"`
	Machine string             `json:"machine"`
	Boot    VirttemplateBoot   `json:"boot"`
	Memory  VirttemplateMemory `json:"memory"`

	CPU VirttemplateCPU `json:"cpu"`

	Topology VirttemplateTopology `json:"topology"`

	Devices VirttemplateDeviceList `json:"devices"`
}

type VirttemplateBoot struct {
	// 'direct' or 'firmware'
	Type string `json:"type"`

	// Only if Type == 'direct'
	Kernel     *VirtStorageVol `json:"kernel,omitempty"`
	Ramdisk    *VirtStorageVol `json:"ramdisk,omitempty"`
	KernelArgs string          `json:"kernel_args,omitempty"`

	Firmware *VirttemplateFirmware `json:"firmware"`
}

type VirttemplateFirmware struct {
	// 'efi' or 'bios'
	Type string `json:"type,omitempty"`
}

type VirttemplateCPUFeature struct {
	Name string `json:"name"`
	// 'force', 'require', 'optional', 'disable', 'forbid'
	Policy string `json:"policy"`
}

type VirttemplateCPU struct {
	Count    int                      `json:"count"`
	Mode     string                   `json:"string"`
	Model    string                   `json:"string"`
	Features []VirttemplateCPUFeature `json:"features"`
}

type VirttemplateMemory struct {
	// Size of DIMMs currently plugged in MB
	Initial int `json:"initial"`
	// Maximum size to allow hotplug DIMMs in MB
	Maximum int `json:"maximum"`

	// Total number of DIMM slots - must be a
	// divisor of both Present and Maximum
	Slots int `json:"slots"`
}

type VirttemplateTopology struct {
	Nodes   int `json:"nodes,omitempty"`
	Sockets int `json:"sockets,omitempty"`
	Cores   int `json:"cores,omitempty"`
	Threads int `json:"threads,omitempty"`
}

type VirttemplateDeviceList struct {
	Disks    []*VirttemplateDisk    `json:"disk"`
	Consoles []*VirttemplateConsole `json:"console"`
	Video    []*VirttemplateVideo   `json:"video"`
}

type VirttemplateDiskEncrypt struct {
	Passphrase string `json:"passphrase"`
}

type VirttemplateDisk struct {
	// 'disk', 'cdrom', etc
	Device    string                   `json:"type"`
	Source    *VirtStorageVol          `json:"source"`
	BootIndex int                      `json:"bootindex"`
	Encrypt   *VirttemplateDiskEncrypt `json:"encrypt"`
}

type VirttemplateConsole struct {
	// 'serial', 'virtio'
	Type string `json:"type"`
}

type VirttemplateVideo struct {
	// 'vga', 'cirrus', 'qxl', 'virtio', 'vmvga'
	Type string `json:"type"`
	VRam int    `json:"vram"`
}
