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
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Virtmachine defines a Virtmachine deployment.
type Virtmachine struct {
	v1.TypeMeta `json:",inline"`
	Metadata    v1.ObjectMeta     `json:"metadata"`
	Spec        VirtmachineSpec   `json:"spec"`
	Status      VirtmachineStatus `json:"status"`
}

// VirtmachineList is a list of Virtmachines.
type VirtmachineList struct {
	v1.TypeMeta `json:",inline"`
	Metadata    v1.ListMeta `json:"metadata"`

	Items []*Virtmachine `json:"items"`
}

// VirtmachineSpec holds specification parameters of a Virtmachine deployment.
type VirtmachineSpec struct {
	// The hardware desired to be applied the running instance
	Hardware VirtmachineHardware `json:"hardware"`
}

type VirtmachineStatus struct {
	// The hardware currently applied to the running instance
	Hardware VirtmachineHardware `json:"hardware"`
}

type VirtmachineHardware struct {
	// Hypervisor type (libvirt: /domain/@type)
	Type    string            `json:"type"`
	Arch    string            `json:"arch"`
	Machine string            `json:"machine"`
	Boot    VirtmachineBoot   `json:"boot"`
	Memory  VirtmachineMemory `json:"memory"`

	CPU VirtmachineCPU `json:"cpu"`

	Topology VirtmachineTopology `json:"topology"`

	Devices VirtmachineDeviceList `json:"devices"`
}

type VirtmachineStorage struct {
	PersistentVolume *VirtmachineStoragePersistentVolume `json:"persistentVolume"`
	ImageFile        *VirtmachineStorageImageFile        `json:"imageFile"`
}

// The guest will be directly connected to the raw persistent storage
// volume listed, assuming QEMU has a network client for the storage
// protocol refered to.
type VirtmachineStoragePersistentVolume struct {
	ClaimName string `json:"claimName"`
}

// The guest will use a local image file associated with resource
// whose k8s name is 'FileName' - nb this is *not* file path on
// disk - this is the TPR resource name
type VirtmachineStorageImageFile struct {
	FileName string `json:"fileName"`
}

type VirtmachineBoot struct {
	// 'direct' or 'firmware'
	Type string `json:"type"`

	// Only if Type == 'direct'
	Kernel     *VirtmachineStorage `json:"kernel,omitempty"`
	Ramdisk    *VirtmachineStorage `json:"ramdisk,omitempty"`
	KernelArgs string              `json:"kernel_args,omitempty"`

	Firmware *VirtmachineFirmware `json:"firmware"`
}

type VirtmachineFirmware struct {
	// 'efi' or 'bios'
	Type string `json:"type,omitempty"`
}

type VirtmachineCPUFeature struct {
	Name string `json:"name"`
	// 'force', 'require', 'optional', 'disable', 'forbid'
	Policy string `json:"policy"`
}

type VirtmachineCPU struct {
	Count    int                     `json:"count"`
	Mode     string                  `json:"string"`
	Model    string                  `json:"string"`
	Features []VirtmachineCPUFeature `json:"features"`
}

type VirtmachineMemory struct {
	// Size of DIMMs currently plugged in MB
	Initial int `json:"initial"`
	// Maximum size to allow hotplug DIMMs in MB
	Maximum int `json:"maximum"`

	// Total number of DIMM slots - must be a
	// divisor of both Present and Maximum
	Slots int `json:"slots"`
}

type VirtmachineTopology struct {
	Nodes   int `json:"nodes,omitempty"`
	Sockets int `json:"sockets,omitempty"`
	Cores   int `json:"cores,omitempty"`
	Threads int `json:"threads,omitempty"`
}

type VirtmachineDeviceList struct {
	Disks    []*VirtmachineDisk    `json:"disk"`
	Consoles []*VirtmachineConsole `json:"console"`
	Video    []*VirtmachineVideo   `json:"video"`
}

type VirtmachineDiskEncrypt struct {
	Passphrase string `json:"passphrase"`
}

type VirtmachineDiskSource struct {
	Name        string `json:"name"`
	BackingName string `json:"name"`
}

type VirtmachineDisk struct {
	// 'disk', 'cdrom', etc
	Device    string                  `json:"device"`
	Source    *VirtmachineStorage     `json:"source"`
	BootIndex int                     `json:"bootindex"`
	Encrypt   *VirtmachineDiskEncrypt `json:"encrypt"`
}

type VirtmachineConsole struct {
	// 'serial', 'virtio'
	Type string `json:"type"`
}

type VirtmachineVideo struct {
	// 'vga', 'cirrus', 'qxl', 'virtio', 'vmvga'
	Type string `json:"type"`
	VRam int    `json:"vram"`
}

// Required to satisfy Object interface
func (ni *Virtmachine) GetObjectKind() schema.ObjectKind {
	return &ni.TypeMeta
}

// Required to satisfy ObjectMetaAccessor interface
func (ni *Virtmachine) GetObjectMeta() v1.Object {
	return &ni.Metadata
}

// Required to satisfy Object interface
func (ni *VirtmachineList) GetObjectKind() schema.ObjectKind {
	return &ni.TypeMeta
}

// Required to satisfy ListMetaAccessor interface
func (ni *VirtmachineList) GetListMeta() v1.List {
	return &ni.Metadata
}
