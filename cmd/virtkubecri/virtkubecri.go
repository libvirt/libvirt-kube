package main

import (
	"flag"
	"fmt"
	"libvirt.org/libvirt-kube/pkg/kubecri"
	"os"
)

var (
	listen = flag.String("listen", "/run/libvirt/virtkubecri.sock",
		"UNIX socket path to listen on")
	connect = flag.String("connect", "qemu:///system",
		"Libvirt connection URI")
)

func main() {
	flag.Parse()

	svc, err := kubecri.NewService(*listen, *connect)
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