package designer

import (
	"fmt"
	"net"

	"github.com/libvirt/libvirt-go-xml"
	"github.com/twinj/uuid"
	"k8s.io/client-go/kubernetes"
	kubeapi "k8s.io/client-go/pkg/api"
	kubeapiv1 "k8s.io/client-go/pkg/api/v1"

	"libvirt.org/libvirt-kube/pkg/api"
	apiv1 "libvirt.org/libvirt-kube/pkg/api/v1alpha1"
)

type DomainDesigner struct {
	clientset *kubernetes.Clientset
	Domain    *libvirtxml.Domain
}

func NewDomainDesigner(clientset *kubernetes.Clientset) *DomainDesigner {
	uuid := uuid.NewV4().String()
	name := fmt.Sprintf("kube-%s", uuid)

	return &DomainDesigner{
		clientset: clientset,
		Domain: &libvirtxml.Domain{
			UUID: uuid,
			Name: name,
		},
	}
}

func (d *DomainDesigner) getVolumeLocalPath(etype string, src *apiv1.VirttemplateStorage) (string, error) {
	return "", fmt.Errorf("Cannot setup volume")
}

func (d *DomainDesigner) setOSConfig(tmpl *apiv1.VirttemplateSpec) error {
	d.Domain.OS = &libvirtxml.DomainOS{
		Type: &libvirtxml.DomainOSType{
			Arch: tmpl.Arch,
			Type: "hvm",
		},
	}

	if tmpl.Machine != "" {
		d.Domain.OS.Type.Machine = tmpl.Machine
	}

	switch tmpl.Boot.Type {
	case "direct":
		kpath, err := d.getVolumeLocalPath("kernel", tmpl.Boot.Kernel)
		if err != nil {
			return err
		}
		ipath, err := d.getVolumeLocalPath("ramdisk", tmpl.Boot.Ramdisk)
		if err != nil {
			return err
		}

		d.Domain.OS.Kernel = kpath
		d.Domain.OS.Initrd = ipath
		d.Domain.OS.KernelArgs = tmpl.Boot.KernelArgs

	case "firmware":
		// We set boot index on devices later for this

	default:
		return fmt.Errorf("Unknown boot type '%s'", tmpl.Boot.Type)
	}

	if tmpl.Boot.Firmware != nil {
		switch tmpl.Boot.Firmware.Type {
		case "efi":
			switch tmpl.Arch {
			case "x86_64":
				d.Domain.OS.Loader = &libvirtxml.DomainLoader{
					Path: "/usr/share/OVMF/OVMF_CODE.fd",
				}
			case "aarch64":
				d.Domain.OS.Loader = &libvirtxml.DomainLoader{
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

func (d *DomainDesigner) setMemoryConfig(tmpl *apiv1.VirttemplateSpec) error {
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

	d.Domain.Memory = &libvirtxml.DomainMemory{
		Value: tmpl.Memory.Initial,
		Unit:  "MiB",
	}
	if 0 == 1 {
		d.Domain.MaximumMemory = &libvirtxml.DomainMaxMemory{
			Value: tmpl.Memory.Initial,
			Unit:  "MiB",
			Slots: futureSlots,
		}
	}

	return nil
}

func (d *DomainDesigner) setCPUConfig(tmpl *apiv1.VirttemplateSpec) error {
	// TODO
	return nil
}

func (d *DomainDesigner) setDiskConfigRBD(src *kubeapiv1.RBDVolumeSource, disk *libvirtxml.DomainDisk) error {
	disk.Type = "network"

	disk.Source = &libvirtxml.DomainDiskSource{
		Protocol: "rbd",
		Name:     src.RBDPool + "/" + src.RBDImage,
	}

	for _, mon := range src.CephMonitors {
		host, port, err := net.SplitHostPort(mon)
		if err != nil {
			return err
		}
		disk.Source.Hosts = append(disk.Source.Hosts,
			libvirtxml.DomainDiskSourceHost{
				Transport: "tcp",
				Name:      host,
				Port:      port,
			})
	}
	disk.Target = &libvirtxml.DomainDiskTarget{
		Dev: "vda",
		Bus: "virtio",
	}

	return nil
}

func (d *DomainDesigner) setDiskConfigISCSI(src *kubeapiv1.ISCSIVolumeSource, disk *libvirtxml.DomainDisk) error {
	disk.Type = "network"

	disk.Source = &libvirtxml.DomainDiskSource{
		Protocol: "rbd",
		Name:     fmt.Sprintf("%s/%d", src.IQN, src.Lun),
	}

	host, port, err := net.SplitHostPort(src.TargetPortal)
	if err != nil {
		return err
	}
	disk.Source.Hosts = append(disk.Source.Hosts,
		libvirtxml.DomainDiskSourceHost{
			Transport: "tcp",
			Name:      host,
			Port:      port,
		})

	disk.Target = &libvirtxml.DomainDiskTarget{
		Dev: "vda",
		Bus: "virtio",
	}

	return nil
}

func (d *DomainDesigner) setDiskConfig(tmpl *apiv1.VirttemplateSpec, disk *apiv1.VirttemplateDisk, devs *libvirtxml.DomainDeviceList) error {
	pvname, pvspec, err := api.GetVolumeSpec(d.clientset, disk.Source.PersistentVolume.ClaimName, kubeapi.NamespaceDefault)
	if err != nil {
		return err
	}

	diskConfig := libvirtxml.DomainDisk{
		Device: disk.Device,
	}

	src := pvspec.PersistentVolumeSource

	if src.RBD != nil {
		err = d.setDiskConfigRBD(src.RBD, &diskConfig)
	} else if src.ISCSI != nil {
		err = d.setDiskConfigISCSI(src.ISCSI, &diskConfig)
	} else {
		return fmt.Errorf("Unsupported persistent volume source on %s", pvname)
	}

	devs.Disks = append(devs.Disks, diskConfig)

	return nil
}

func (d *DomainDesigner) setDeviceConfig(tmpl *apiv1.VirttemplateSpec) error {
	d.Domain.Devices = &libvirtxml.DomainDeviceList{}

	for _, disk := range tmpl.Devices.Disks {
		if err := d.setDiskConfig(tmpl, disk, d.Domain.Devices); err != nil {
			return err
		}
	}

	return nil
}

func (d *DomainDesigner) SetResourcePartition(partition string) {
	d.Domain.Resource = &libvirtxml.DomainResource{
		Partition: partition,
	}
}

func (d *DomainDesigner) ApplyVirtTemplate(tmpl *apiv1.VirttemplateSpec) error {
	d.Domain.Type = tmpl.Type

	if err := d.setOSConfig(tmpl); err != nil {
		return err
	}

	if err := d.setMemoryConfig(tmpl); err != nil {
		return err
	}

	if err := d.setCPUConfig(tmpl); err != nil {
		return err
	}

	if err := d.setDeviceConfig(tmpl); err != nil {
		return err
	}

	return nil
}
