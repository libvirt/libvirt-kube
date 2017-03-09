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

package vmangel

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ghodss/yaml"
	"github.com/golang/glog"
	"golang.org/x/net/context"
	kubeapi "k8s.io/client-go/pkg/api"

	"libvirt.org/libvirt-kube/pkg/vmshim/rpc"
)

type Guardian struct {
	machine     string
	namespace   string
	pod         string
	context     *context.Context
	shimAddr    string
	shimTimeout time.Duration
	sighandler  chan os.Signal
}

func NewGuardian(machine, namespace string, shimAddr string, shimTimeout time.Duration) (*Guardian, error) {

	if namespace == "" {
		namespace = kubeapi.NamespaceDefault
	}
	if machine == "" {
		return nil, fmt.Errorf("Machine name cannot be empty")
	}

	sighandler := make(chan os.Signal, 1)

	signal.Notify(sighandler, syscall.SIGHUP)
	signal.Notify(sighandler, syscall.SIGTERM)
	signal.Notify(sighandler, syscall.SIGINT)
	signal.Notify(sighandler, syscall.SIGQUIT)

	return &Guardian{
		shimAddr:    shimAddr,
		shimTimeout: shimTimeout,
		machine:     machine,
		namespace:   namespace,
		sighandler:  sighandler,
	}, nil
}

func (g *Guardian) waitForError(shimconn net.Conn, notify chan error) {
	msg := make([]byte, 1024)
	n, err := shimconn.Read(msg)
	if err != nil {
		notify <- err
	}
	if n == 0 {
		notify <- fmt.Errorf("Machine terminated without error message")
	} else {
		notify <- fmt.Errorf("%s", string(msg[0:n]))
	}
}

func (g *Guardian) Watch() error {
	shimconn, err := net.DialTimeout("unix", g.shimAddr, g.shimTimeout)
	if err != nil {
		return err
	}
	defer shimconn.Close()

	glog.V(1).Infof("Requesting start %s/%s", g.namespace, g.machine)
	info := &rpc.MachineStartInfo{
		Pod:       g.pod,
		Machine:   g.machine,
		Namespace: g.namespace,
	}
	infobuf, err := yaml.Marshal(&info)
	if err != nil {
		return nil
	}

	n, err := shimconn.Write(infobuf)
	if err != nil {
		return nil
	}
	if n != len(infobuf) {
		return fmt.Errorf("Short write to shim")
	}
	// Leave shimconn open hereafter - closing the socket is
	// the sign the shim uses to kill off the VM

	errnotify := make(chan error)
	go g.waitForError(shimconn, errnotify)

	glog.V(1).Infof("Started %s/%s", g.namespace, g.machine)

	for {
		select {
		case sig := <-g.sighandler:
			glog.V(1).Infof("Signal %s, stopping wait", sig)
			return nil
		case err := <-errnotify:
			glog.V(1).Infof("Machine done %s", err)
			return nil
		}
	}
}
