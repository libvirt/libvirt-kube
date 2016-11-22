package main

import (
	"libvirt.org/libvirt-kubelet/pkg/service"
	"os"
)

func main() {

	svc, err := service.New("/run/libvirt-kubelet.sock")
	if err != nil {
		os.Exit(1)
	}

	err = svc.Run()
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
