package designer

import (
	"errors"
	"fmt"
	"github.com/libvirt/libvirt-go-xml"
	"github.com/twinj/uuid"
	kubeapi "libvirt.org/libvirt-kube/pkg/kubeapi/v1alpha1"
)

func getVolumeLocalPath(vol *kubeapi.VirtStorageVol) (string, error) {
	if vol.Pool.Dir == nil {
		return "", errors.New("A local directory volume is required")
	}

	return vol.Pool.Dir.Path + "/" + vol.Name, nil
}

func setDomainOSConfig(tmpl *kubeapi.VirttemplateSpec, dom *libvirtxml.Domain) error {
	dom.OS = &libvirtxml.DomainOS{
		Type: &libvirtxml.DomainOSType{
			Arch: tmpl.Arch,
			Type: "hvm",
		},
	}

	if tmpl.Machine != "" {
		dom.OS.Type.Machine = tmpl.Machine
	}

	switch tmpl.Boot.Type {
	case "firmware":
		kpath, err := getVolumeLocalPath(tmpl.Boot.Kernel)
		if err != nil {
			return err
		}
		ipath, err := getVolumeLocalPath(tmpl.Boot.Ramdisk)
		if err != nil {
			return err
		}

		dom.OS.Kernel = kpath
		dom.OS.Initrd = ipath
		dom.OS.KernelArgs = tmpl.Boot.KernelArgs

	case "direct":
		// We set boot index on devices later for this

	default:
		return fmt.Errorf("Unknown boot type '%s'", tmpl.Boot.Type)
	}

	if tmpl.Boot.Firmware != nil {
		switch tmpl.Boot.Firmware.Type {
		case "efi":
			switch tmpl.Arch {
			case "x86_64":
				dom.OS.Loader = &libvirtxml.DomainLoader{
					Path: "/usr/share/OVMF/OVMF_CODE.fd",
				}
			case "aarch64":
				dom.OS.Loader = &libvirtxml.DomainLoader{
					Path: "/usr/share/AAVMF/AAVMF_CODE.fd",
				}
			default:
				return fmt.Errorf("Architecture '%s' does not support 'efi' firmware", tmpl.Arch)
			}
		case "bios":
			switch tmpl.Arch {
			case "x86_64":
				// nada, its the default
			case "i686":
				// nada, its the default
			default:
				return fmt.Errorf("Architecture '%s' does not support 'bios' firmware", tmpl.Arch)
			}
		default:
			return fmt.Errorf("Unknown firmware type '%s'", tmpl.Boot.Firmware.Type)
		}
	}

	return nil
}

func setDomainMemoryConfig(tmpl *kubeapi.VirttemplateSpec, dom *libvirtxml.Domain) error {
	if (tmpl.Memory.Initial%tmpl.Memory.Slots) != 0 ||
		(tmpl.Memory.Maximum%tmpl.Memory.Slots) != 0 {
		return fmt.Errorf("Memory present %d and maximum %d must be multiple of slots %d",
			tmpl.Memory.Initial,
			tmpl.Memory.Maximum,
			tmpl.Memory.Slots)
	}

	slotSize := tmpl.Memory.Maximum / tmpl.Memory.Slots
	initialSlots := tmpl.Memory.Initial / slotSize
	futureSlots := tmpl.Memory.Slots - initialSlots

	dom.Memory = &libvirtxml.DomainMemory{
		Value: tmpl.Memory.Initial,
		Unit:  "MiB",
	}
	if 0 == 1 {
		dom.MaximumMemory = &libvirtxml.DomainMaxMemory{
			Value: tmpl.Memory.Initial,
			Unit:  "MiB",
			Slots: futureSlots,
		}
	}

	return nil
}

func setDomainCPUConfig(tmpl *kubeapi.VirttemplateSpec, dom *libvirtxml.Domain) error {
	// TODO
	return nil
}

func setDomainDiskConfig(tmpl *kubeapi.VirttemplateSpec, disk *kubeapi.VirttemplateDisk, devs *libvirtxml.DomainDeviceList) error {
	path, err := getVolumeLocalPath(disk.Source)
	if err != nil {
		return err
	}

	diskConfig := libvirtxml.DomainDisk{
		Type:   "file",
		Device: disk.Device,
		FileSource: &libvirtxml.DomainDiskFileSource{
			File: path,
		},
	}

	devs.Disks = append(devs.Disks, diskConfig)

	return nil
}

func setDomainDeviceConfig(tmpl *kubeapi.VirttemplateSpec, dom *libvirtxml.Domain) error {
	dom.Devices = &libvirtxml.DomainDeviceList{}

	for _, disk := range tmpl.Devices.Disks {
		if err := setDomainDiskConfig(tmpl, disk, dom.Devices); err != nil {
			return err
		}
	}

	return nil
}

func DomainConfigFromVirtTemplate(tmpl *kubeapi.VirttemplateSpec, partition string) (*libvirtxml.Domain, error) {
	uuid := uuid.NewV4().String()
	name := fmt.Sprintf("kube-%s", uuid)

	dom := &libvirtxml.Domain{
		Type: tmpl.Type,
		UUID: uuid,
		Name: name,
	}

	if partition != "" {
		dom.Resource = &libvirtxml.DomainResource{
			Partition: partition,
		}
	}

	if err := setDomainOSConfig(tmpl, dom); err != nil {
		return nil, err
	}

	if err := setDomainMemoryConfig(tmpl, dom); err != nil {
		return nil, err
	}

	if err := setDomainCPUConfig(tmpl, dom); err != nil {
		return nil, err
	}

	return dom, nil
}
