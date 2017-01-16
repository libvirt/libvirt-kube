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
	// Prefered values
	Nodes   int `json:"nodes,omitempty"`
	Sockets int `json:"sockets,omitempty"`
	Cores   int `json:"cores,omitempty"`
	Threads int `json:"threads,omitempty"`

	// Absolute minimum values
	MinNodes   int `json:"min_nodes,omitempty"`
	MinSockets int `json:"min_sockets,omitempty"`
	MinCores   int `json:"min_cores,omitempty"`
	MinThreads int `json:"min_threads,omitempty"`

	// Absolute maximum values
	MaxNodes   int `json:"max_nodes,omitempty"`
	MaxSockets int `json:"max_sockets,omitempty"`
	MaxCores   int `json:"max_cores,omitempty"`
	MaxThreads int `json:"max_threads,omitempty"`
}

type VirtTemplateDeviceList struct {
	Disks []*VirtTemplateDisk `json:"disk"`
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
