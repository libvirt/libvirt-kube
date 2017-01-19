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

// VirtTemplate defines a VirtTemplate deployment.
type VirtTemplate struct {
	metav1.TypeMeta `json:",inline"`
	v1.ObjectMeta   `json:"metadata,omitempty"`
	Spec            VirtTemplateSpec `json:"spec"`
	//	Status          *VirtTemplateStatus `json:"status,omitempty"`
}

// VirtTemplateList is a list of VirtTemplates.
type VirtTemplateList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []*VirtTemplate `json:"items"`
}

// VirtTemplateSpec holds specification parameters of a VirtTemplate deployment.
type VirtTemplateSpec struct {
	// Hypervisor type (libvirt: /domain/@type)
	Type    string             `json:"type"`
	Arch    string             `json:"arch"`
	Machine string             `json:"machine"`
	Boot    VirtTemplateBoot   `json:"boot"`
	Memory  VirtTemplateMemory `json:"memory"`

	CPU VirtTemplateCPU `json:"cpu"`

	Topology VirtTemplateTopology `json:"topology"`

	Devices VirtTemplateDeviceList `json:"devices"`
}

type VirtTemplateBoot struct {
	// 'direct' or 'firmware'
	Type string `json:"type"`

	// Only if Type == 'direct'
	Kernel     *VirtStorageVol `json:"kernel,omitempty"`
	Ramdisk    *VirtStorageVol `json:"ramdisk,omitempty"`
	KernelArgs string          `json:"kernel_args,omitempty"`

	Firmware *VirtTemplateFirmware `json:"firmware"`
}

type VirtTemplateFirmware struct {
	// 'efi' or 'bios'
	Type string `json:"type,omitempty"`
}

type VirtTemplateCPUFeature struct {
	Name string `json:"name"`
	// 'force', 'require', 'optional', 'disable', 'forbid'
	Policy string `json:"policy"`
}

type VirtTemplateCPU struct {
	Count    int                      `json:"count"`
	Mode     string                   `json:"string"`
	Model    string                   `json:"string"`
	Features []VirtTemplateCPUFeature `json:"features"`
}

type VirtTemplateMemory struct {
	// Size of DIMMs currently plugged in MB
	Initial int `json:"initial"`
	// Maximum size to allow hotplug DIMMs in MB
	Maximum int `json:"maximum"`

	// Total number of DIMM slots - must be a
	// divisor of both Present and Maximum
	Slots int `json:"slots"`
}

type VirtTemplateTopology struct {
	Nodes   int `json:"nodes,omitempty"`
	Sockets int `json:"sockets,omitempty"`
	Cores   int `json:"cores,omitempty"`
	Threads int `json:"threads,omitempty"`
}

type VirtTemplateDeviceList struct {
	Disks    []*VirtTemplateDisk    `json:"disk"`
	Consoles []*VirtTemplateConsole `json:"console"`
	Video    []*VirtTemplateVideo   `json:"video"`
}

type VirtTemplateDiskEncryptLUKS struct {
	// Name of kubernetes secret
	Passphrase string `json:"passphrase"`
}

type VirtTemplateDiskEncrypt struct {
	LUKS *VirtTemplateDiskEncryptLUKS `json:"luks"`
}

type VirtTemplateDisk struct {
	// 'disk', 'cdrom', etc
	Device    string                       `json:"type"`
	Source    *VirtStorageVol              `json:"source"`
	BootIndex int                          `json:"bootindex"`
	Encrypt   *VirtTemplateDiskEncryptLUKS `json:"encrypt"`
}

type VirtTemplateConsole struct {
	// 'serial', 'virtio'
	Type string `json:"type"`
}

type VirtTemplateVideo struct {
	// 'vga', 'cirrus', 'qxl', 'virtio', 'vmvga'
	Type string `json:"type"`
	VRam int    `json:"vram"`
}
