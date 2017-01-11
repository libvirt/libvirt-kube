package main

import (
	"flag"
	"fmt"
	"libvirt.org/libvirt-kubelet/pkg/service"
	"os"
)

var (
	listen = flag.String("listen", "/run/libvirt-kubelet.sock",
		"UNIX socket path to listen on")
)

func main() {
	flag.Parse()

	svc, err := service.New(*listen)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = svc.Run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
