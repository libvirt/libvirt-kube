package config

import (
	"encoding/xml"
	"strings"
	"testing"
)

var testData = []struct {
	Object   *Domain
	Expected []string
}{
	{
		Object: &Domain{
			Type: "kvm",
			Name: "test",
		},
		Expected: []string{
			`<domain type="kvm">`,
			`  <name>test</name>`,
			`</domain>`,
		},
	},
	{
		Object: &Domain{
			Type: "kvm",
			Name: "test",
			Devices: &DomainDeviceList{
				Disks: []DomainDisk{
					DomainDisk{
						Type:   "file",
						Device: "cdrom",
						Driver: DomainDiskDriver{
							Name: "qemu",
							Type: "qcow2",
						},
						FileSource: DomainDiskFileSource{
							File: "/var/lib/libvirt/images/demo.qcow2",
						},
					},
				},
			},
		},
		Expected: []string{
			`<domain type="kvm">`,
			`  <name>test</name>`,
			`  <devices>`,
			`    <disk type="file" device="cdrom">`,
			`      <driver name="qemu" type="qcow2"></driver>`,
			`      <source file="/var/lib/libvirt/images/demo.qcow2"></source>`,
			`    </disk>`,
			`  </devices>`,
			`</domain>`,
		},
	},
	{
		Object: &Domain{
			Type: "kvm",
			Name: "test",
			Devices: &DomainDeviceList{
				Inputs: []DomainInput{
					DomainInput{
						Type: "tablet",
						Bus:  "usb",
					},
					DomainInput{
						Type: "keyboard",
						Bus:  "ps2",
					},
				},
				Videos: []DomainVideo{
					DomainVideo{
						Model: DomainVideoModel{
							Type: "cirrus",
						},
					},
				},
				Graphics: []DomainGraphic{
					DomainGraphic{
						Type: "vnc",
					},
				},
			},
		},
		Expected: []string{
			`<domain type="kvm">`,
			`  <name>test</name>`,
			`  <devices>`,
			`    <input type="tablet" bus="usb"></input>`,
			`    <input type="keyboard" bus="ps2"></input>`,
			`    <graphics type="vnc"></graphics>`,
			`    <video>`,
			`      <model type="cirrus"></model>`,
			`    </video>`,
			`  </devices>`,
			`</domain>`,
		},
	},
}

func TestDomain(t *testing.T) {
	for _, test := range testData {
		doc, err := xml.MarshalIndent(test.Object, "", "  ")
		if err != nil {
			t.Fatal(err)
		}

		expect := strings.Join(test.Expected, "\n")

		if string(doc) != expect {
			t.Fatal("Bad xml:\n", string(doc), "\n does not match\n", expect, "\n")
		}
	}
}
