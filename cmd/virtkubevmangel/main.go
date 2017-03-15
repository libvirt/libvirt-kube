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

package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/spf13/pflag"
	kubeapi "k8s.io/client-go/pkg/api"

	"libvirt.org/libvirt-kube/pkg/vmangel"
)

var (
	machine   = pflag.String("machine", "", "Name of virtmachine resource")
	pod       = pflag.String("pod", "", "Name of POD associated with the machine")
	namespace = pflag.String("namespace", "", "Namespace in which machine and pod resource were created")
	shimaddr  = pflag.String("shimaddr", "/var/run/virtkubevmshim/shim.sock",
		"UNIX socket path for virtkubvmshim server")
)

func main() {
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	// Convince glog that we really have parsed CLI
	flag.CommandLine.Parse([]string{})

	if *namespace == "" {
		*namespace = os.Getenv("LIBVIRT_KUBE_VM_ANGEL_NAMESPACE")
		if *namespace == "" {
			*namespace = kubeapi.NamespaceDefault
		}
	}

	if *pod == "" {
		*pod = os.Getenv("LIBVIRT_KUBE_VM_ANGEL_POD")
		if *pod == "" {
			fmt.Println("Pod name cannot be empty")
			os.Exit(1)
		}
	}

	if *machine == "" {
		fmt.Println("Machine name cannot be empty")
		os.Exit(1)
	}

	gdn, err := vmangel.NewGuardian(*machine, *namespace, *pod, *shimaddr, 5*time.Second)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	err = gdn.Watch()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(0)
}
