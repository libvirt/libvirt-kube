package nodeinfo

import (
	"encoding/json"
	"github.com/libvirt/libvirt-go"
	"github.com/libvirt/libvirt-go-xml"
)

type NodeInfoGuest struct {
	Hypervisor string
	Arch       string
	Type       string

	Machines []string
}

type NodeInfoMemory struct {
	PageSize int
	Present  int
	Used     int
}

type NodeInfoNUMACell struct {
	CPUSAvail int
	CPUSUsed  int
	Memory    []NodeInfoMemory
}

type NodeInfoResources struct {
	NUMACells []NodeInfoNUMACell
}

type NodeInfo struct {
	UUID      string
	Arch      string
	Guests    []NodeInfoGuest
	Resources NodeInfoResources
}

func NewNodeInfo(caps *libvirtxml.Caps, conn *libvirt.Connect) (*NodeInfo, error) {
	guests := make([]NodeInfoGuest, 0)

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
			guests = append(guests, NodeInfoGuest{
				Hypervisor: cdom.Type,
				Arch:       cguest.Arch.Name,
				Type:       cguest.OSType,
				Machines:   machines,
			})
		}
	}

	cells := make([]NodeInfoNUMACell, 0)
	if caps.Host.NUMA != nil {
		for _, lvcell := range caps.Host.NUMA.Cells {
			ncpus := len(lvcell.CPUS)
			memory := make([]NodeInfoMemory, 0)
			if lvcell.PageInfo == nil {
				memory = append(memory, NodeInfoMemory{
					PageSize: 4096,
				})
			} else {
				for _, lvpage := range lvcell.PageInfo {
					memory = append(memory, NodeInfoMemory{
						PageSize: lvpage.Size,
					})
				}
			}
			cells = append(cells, NodeInfoNUMACell{
				CPUSAvail: ncpus,
				CPUSUsed:  0,
				Memory:    memory,
			})
		}
	} else {
		nodeinfo, err := conn.GetNodeInfo()
		if err != nil {
			return nil, err
		}
		ncpus := int(nodeinfo.Nodes * nodeinfo.Sockets * nodeinfo.Cores * nodeinfo.Threads)
		memory := make([]NodeInfoMemory, 0)
		memory = append(memory, NodeInfoMemory{
			PageSize: 4096,
		})
		cells = append(cells, NodeInfoNUMACell{
			CPUSAvail: ncpus,
			CPUSUsed:  0,
			Memory:    memory,
		})
	}

	resources := NodeInfoResources{
		NUMACells: cells,
	}

	info := &NodeInfo{
		UUID:      caps.Host.UUID,
		Arch:      caps.Host.CPU.Arch,
		Guests:    guests,
		Resources: resources,
	}

	return info, nil
}

func (n *NodeInfo) Serialize() (string, error) {

	data, err := json.Marshal(n)

	return string(data), err
}

func (n *NodeInfo) Deserialize(data string) error {

	err := json.Unmarshal([]byte(data), n)

	return err
}
