package config

import (
	"encoding/xml"
)

type DomainController struct {
	Type  string `xml:"type,attr"`
	Index string `xml:"index,attr"`
}

type DomainDiskFileSource struct {
	File string `xml:"file,attr"`
}

type DomainDiskDriver struct {
	Name string `xml:"name,attr"`
	Type string `xml:"type,attr"`
}

type DomainDisk struct {
	Type       string               `xml:"type,attr"`
	Device     string               `xml:"device,attr"`
	Driver     DomainDiskDriver     `xml:"driver"`
	FileSource DomainDiskFileSource `xml:"source"`
}

type DomainInterfaceMAC struct {
	Address string `xml:"address,attr"`
}

type DomainInterfaceModel struct {
	Type string `xml:type,attr"`
}

type DomainInterface struct {
	Type  string               `xml:"type,attr"`
	MAC   string               `xml:"mac"`
	Model DomainInterfaceModel `xml:"model"`
}

type DomainChardev struct {
	Type string `xml:"type,attr"`
}

type DomainInput struct {
	Type string `xml:"type,attr"`
	Bus  string `xml:"bus,attr"`
}

type DomainGraphic struct {
	Type string `xml:"type,attr"`
}

type DomainVideoModel struct {
	Type string `xml:"type,attr"`
}

type DomainVideo struct {
	Model DomainVideoModel `xml:"model"`
}

type DomainDeviceList struct {
	Controllers []DomainController `xml:"controller"`
	Disks       []DomainDisk       `xml:"disk"`
	Interfaces  []DomainInterface  `xml:"interface"`
	Serials     []DomainChardev    `xml:"serial"`
	Consoles    []DomainChardev    `xml:"console"`
	Inputs      []DomainInput      `xml:"input"`
	Graphics    []DomainGraphic    `xml:"graphics"`
	Videos      []DomainVideo      `xml:"video"`
}

type DomainMemory struct {
	Value string `xml:",chardata"`
	Unit  string `xml:"unit,attr"`
}

type Domain struct {
	XMLName       xml.Name          `xml:"domain"`
	Type          string            `xml:"type,attr"`
	Name          string            `xml:"name"`
	UUID          *string           `xml:"uuid"`
	Memory        *DomainMemory     `xml:"memory"`
	CurrentMemory *DomainMemory     `xml:"currentMemory"`
	Devices       *DomainDeviceList `xml:"devices"`
}
