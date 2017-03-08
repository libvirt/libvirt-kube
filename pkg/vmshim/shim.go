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

package vmshim

import (
	"fmt"
	"net"
	"os"
	"path"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/ghodss/yaml"
	"github.com/golang/glog"
	"github.com/libvirt/libvirt-go"
	"k8s.io/client-go/kubernetes"
	kubeapi "k8s.io/client-go/pkg/api"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"libvirt.org/libvirt-kube/pkg/api"
	"libvirt.org/libvirt-kube/pkg/designer"
	"libvirt.org/libvirt-kube/pkg/libvirtutil"
	"libvirt.org/libvirt-kube/pkg/resource"
	"libvirt.org/libvirt-kube/pkg/vmshim/rpc"
)

type Shim struct {
	shimAddr        string
	clientset       *kubernetes.Clientset
	kubeconfig      *rest.Config
	imageRepoPath   string
	imageRepoClient *api.VirtimagerepoClient
	imageFileClient *api.VirtimagefileClient
	conn            *libvirt.Connect
	domains         map[string]*libvirt.Domain
	connNotify      chan libvirtutil.ConnectEvent
	shutdown        map[string]chan bool
	lock            sync.Mutex
	skipValidate    bool
}

func getKubeConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func NewShim(shimAddr string, skipValidate bool, libvirtURI string, imageRepoPath string, kubeconfigfile string) (*Shim, error) {
	kubeconfig, err := getKubeConfig(kubeconfigfile)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	imageRepoClient, err := api.NewVirtimagerepoClient(kubeapi.NamespaceDefault, kubeconfig)
	if err != nil {
		return nil, err
	}

	imageFileClient, err := api.NewVirtimagefileClient(kubeapi.NamespaceDefault, kubeconfig)
	if err != nil {
		return nil, err
	}

	shim := &Shim{
		skipValidate:    skipValidate,
		shimAddr:        shimAddr,
		kubeconfig:      kubeconfig,
		clientset:       clientset,
		imageRepoPath:   imageRepoPath,
		imageFileClient: imageFileClient,
		imageRepoClient: imageRepoClient,
		connNotify:      make(chan libvirtutil.ConnectEvent, 1),
		shutdown:        make(map[string]chan bool),
		domains:         make(map[string]*libvirt.Domain),
	}

	libvirtutil.OpenConnect(libvirtURI, shim.connNotify)

	return shim, nil
}

func (s *Shim) runClientImpl(conn net.Conn) error {
	uconn, ok := conn.(*net.UnixConn)
	if !ok {
		return fmt.Errorf("Expected a UNIX socket but got %s", conn)
	}

	partition := ""
	if s.skipValidate {
		glog.V(1).Infof("Skipping client validation, insecure")
	} else {
		fconn, err := uconn.File()
		if err != nil {
			return fmt.Errorf("Failed to get file for UNIX socket %s: %s", uconn, err)
		}
		defer fconn.Close()

		ucred, err := syscall.GetsockoptUcred(int(fconn.Fd()), syscall.SOL_SOCKET, syscall.SO_PEERCRED)
		if err != nil {
			return fmt.Errorf("Failed to get ucred for socket %d: %s", fconn.Fd(), err)
		}

		glog.V(1).Infof("Peer pid=%d uid=%d gid=%d", ucred.Pid, ucred.Uid, ucred.Gid)

		partition, err = resource.GetResourcePartition(int(ucred.Pid), "name=systemd")
		if err != nil {
			return fmt.Errorf("Failed to get partition for pid %d: %s", ucred.Pid, err)
		}
		glog.V(1).Infof("Using partition %s", partition)

		_, file := path.Split(partition)

		if !strings.HasPrefix(file, "docker-") || !strings.HasSuffix(file, ".scope") {
			return fmt.Errorf("Cgroup '%s' doesn't appear to be from Docker", file)
		}

		containerID := file[6 : len(file)-6]
		glog.V(1).Infof("Docker container ID is %s", containerID)
	}

	infobuf := make([]byte, 1024)
	len, err := conn.Read(infobuf)
	if err != nil {
		return fmt.Errorf("Cannot read info buf")
	}

	info := &rpc.MachineStartInfo{}
	glog.V(1).Infof("Parse '%s'", infobuf)
	err = yaml.Unmarshal(infobuf[0:len], info)
	if err != nil {
		return fmt.Errorf("Cannot unmarshal info buf: %s", err)
	}

	dom, err := s.startMachine(info.Namespace, info.Machine, partition)
	if err != nil {
		return err
	}

	err = s.waitForMachineStop(conn, dom)
	if err != nil {
		return err
	}

	return nil
}

func (s *Shim) runClient(conn net.Conn) {
	defer conn.Close()

	err := s.runClientImpl(conn)
	glog.V(1).Infof("Client error %s", err)
	resp := []byte(fmt.Sprintf("%s", err))
	conn.Write(resp)
}

func (s *Shim) runServer(sock net.Listener) {
	defer sock.Close()

	for {
		glog.V(1).Info("Waiting for client")
		client, err := sock.Accept()
		if err != nil {
			continue
		}

		go s.runClient(client)
	}
}

func (s *Shim) Run() error {
	glog.V(1).Infof("Running on %s", s.shimAddr)
	err := syscall.Unlink(s.shimAddr)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	sock, err := net.Listen("unix", s.shimAddr)
	if err != nil {
		return err
	}

	go s.runServer(sock)

	for {
		select {
		case hypEvent := <-s.connNotify:
			switch hypEvent.Type {
			case libvirtutil.ConnectReady:
				s.lock.Lock()
				glog.V(1).Infof("Setup new connection")
				s.conn = hypEvent.Conn
				s.lock.Unlock()
			case libvirtutil.ConnectFailed:
				s.lock.Lock()
				glog.V(1).Infof("Discard old connection")
				s.conn.Close()
				s.conn = nil
				s.lock.Unlock()
			}
		}
	}

	return nil
}

func (s *Shim) domainLifecycleEvent(c *libvirt.Connect, d *libvirt.Domain, ev *libvirt.DomainEventLifecycle) {
	uuid, err := d.GetUUIDString()
	if err != nil {
		glog.V(1).Infof("Error getting domain UUID", err)
		return
	}
	if ev.Event == libvirt.DOMAIN_EVENT_STOPPED {
		s.lock.Lock()
		shutdown, ok := s.shutdown[uuid]
		s.lock.Unlock()
		glog.V(1).Infof("Notify shutdown domain %s", uuid)
		if ok {
			shutdown <- true
		}
	}
}

func (s *Shim) startMachine(namespace, name, partition string) (*libvirt.Domain, error) {
	glog.V(1).Infof("Start machine='%s', namespace='%s'", name, namespace)

	machineClient, err := api.NewVirtmachineClient(namespace, s.kubeconfig)
	if err != nil {
		return nil, err
	}

	machine, err := machineClient.Get(name)
	if err != nil {
		return nil, err
	}

	domdesign := designer.NewDomainDesigner(s.clientset, s.imageRepoPath, s.imageRepoClient, s.imageFileClient)
	if partition != "" {
		domdesign.SetResourcePartition(partition)
	}
	err = domdesign.ApplyVirtMachine(&machine.Spec)
	if err != nil {
		return nil, err
	}

	cfg := domdesign.Domain

	s.lock.Lock()
	if s.conn == nil {
		s.lock.Unlock()
		return nil, fmt.Errorf("Not currently connected to libvirt")
	}
	conn := s.conn
	conn.Ref()
	s.lock.Unlock()

	defer conn.Close()

	dom, _ := conn.LookupDomainByUUIDString(cfg.UUID)
	if dom != nil {
		dom.Free()
		return nil, fmt.Errorf("Domain with UUID %s already exists", cfg.UUID)
	}
	dom, _ = conn.LookupDomainByName(cfg.Name)
	if dom != nil {
		dom.Free()
		return nil, fmt.Errorf("Domain with UUID %s already exists", cfg.Name)
	}

	for _, secdesign := range domdesign.Secrets {
		secCFG, err := secdesign.Secret.Marshal()
		if err != nil {
			return nil, err
		}
		sec, err := conn.SecretDefineXML(secCFG, 0)
		if err != nil {
			return nil, err
		}
		defer sec.Undefine()

		glog.V(1).Infof("Setting secret %s value", secdesign.Secret.UUID)
		err = sec.SetValue(secdesign.Value, 0)
		if err != nil {
			return nil, err
		}
	}

	cfgXML, err := cfg.Marshal()
	if err != nil {
		return nil, err
	}
	glog.V(1).Infof("Creating domain %s: %s", cfg.UUID, cfgXML)
	domain, err := conn.DomainCreateXML(cfgXML,
		libvirt.DOMAIN_START_AUTODESTROY|libvirt.DOMAIN_START_VALIDATE)
	if err != nil {
		return nil, err
	}

	key := fmt.Sprintf("%s/%s", namespace, name)

	s.domains[key] = domain

	return domain, nil
}

func (s *Shim) waitForClientEOF(conn net.Conn, notify chan bool) {
	data := make([]byte, 1024)

	conn.Read(data)

	notify <- true
}

func (s *Shim) waitForMachineStop(conn net.Conn, dom *libvirt.Domain) error {
	defer func() {
		dom.Destroy()
		dom.Free()
	}()

	uuid, err := dom.GetUUIDString()
	if err != nil {
		return err
	}

	shutdown := make(chan bool, 1)
	s.lock.Lock()
	s.shutdown[uuid] = shutdown
	s.lock.Unlock()

	// State may have changed between starting it and registering
	// for lifecycle events so check again now. After this point
	// we're guaranteed an event
	isActive, err := dom.IsActive()
	if err != nil {
		return err
	}
	eofNotify := make(chan bool, 1)
	go s.waitForClientEOF(conn, eofNotify)
	if isActive {
		// We're running now, so block until we
		// either see the guest shutdown event,
		// or get a signal indicating we should
		// quit.
		select {
		case _ = <-eofNotify:
			glog.V(1).Info("Saw client exit, killing guest")
			s.stopMachine(dom)

		case _ = <-shutdown:
			glog.V(1).Info("Saw guest shutdown, exiting")
			// nada
		}
	} else {
		glog.V(1).Info("Guest already shutdown, exiting")
	}

	s.lock.Lock()
	delete(s.shutdown, uuid)
	s.lock.Unlock()

	return nil
}

func (s *Shim) stopMachine(dom *libvirt.Domain) {
	for {
		err := dom.Destroy()
		if err == nil {
			glog.V(1).Info("Killed guest, exiting")
			return
		}
		lverr, ok := err.(*libvirt.Error)
		if ok && (lverr.Code == libvirt.ERR_NO_DOMAIN || lverr.Code == libvirt.ERR_OPERATION_INVALID) {
			glog.V(1).Info("Guest already quit")
			return
		}
		glog.V(1).Info("Error killing guest %s, retrying", err)
		time.Sleep(250 * time.Millisecond)
	}
}
