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
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/ghodss/yaml"
	"github.com/golang/glog"
	"github.com/libvirt/libvirt-go"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	apiv1 "libvirt.org/libvirt-kube/pkg/api/v1alpha1"
	"libvirt.org/libvirt-kube/pkg/designer"
	"libvirt.org/libvirt-kube/pkg/resource"
)

type Shim struct {
	clientset  *kubernetes.Clientset
	template   *apiv1.VirttemplateSpec
	hypervisor *libvirt.Connect
	domain     *libvirt.Domain
	shutdown   chan bool
	sighandler chan os.Signal
}

func runEventLoop() {
	for {
		libvirt.EventRunDefaultImpl()
	}
}

func init() {
	libvirt.EventRegisterDefaultImpl()
	go runEventLoop()
}

func getKubeConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func NewShim(templateFile string, libvirtURI string, kubeconfigfile string) (*Shim, error) {
	kubeconfig, err := getKubeConfig(kubeconfigfile)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	hypervisor, err := libvirt.NewConnect(libvirtURI)
	if err != nil {
		return nil, err
	}

	templateYAML, err := ioutil.ReadFile(templateFile)
	if err != nil {
		return nil, err
	}

	template := &apiv1.VirttemplateSpec{}
	err = yaml.Unmarshal(templateYAML, &template)
	if err != nil {
		return nil, err
	}

	shim := &Shim{
		clientset:  clientset,
		template:   template,
		hypervisor: hypervisor,
		shutdown:   make(chan bool, 1),
		sighandler: make(chan os.Signal, 1),
	}

	signal.Notify(shim.sighandler, syscall.SIGHUP)
	signal.Notify(shim.sighandler, syscall.SIGTERM)
	signal.Notify(shim.sighandler, syscall.SIGINT)
	signal.Notify(shim.sighandler, syscall.SIGQUIT)

	return shim, nil
}

func (s *Shim) domainLifecycleEvent(c *libvirt.Connect, d *libvirt.Domain, ev *libvirt.DomainEventLifecycle) {
	if ev.Event == libvirt.DOMAIN_EVENT_STOPPED {
		s.shutdown <- true
	}
}

func (s *Shim) Run() error {
	partition, err := resource.GetResourcePartition(os.Getpid())
	if err != nil {
		return err
	}
	glog.V(1).Infof("Using partition %s", partition)

	cfg, err := designer.DomainConfigFromVirtTemplate(s.clientset, s.template, partition)
	if err != nil {
		return err
	}

	dom, _ := s.hypervisor.LookupDomainByUUIDString(cfg.UUID)
	if dom != nil {
		dom.Free()
		return fmt.Errorf("Domain with UUID %s already exists", cfg.UUID)
	}
	dom, _ = s.hypervisor.LookupDomainByName(cfg.Name)
	if dom != nil {
		dom.Free()
		return fmt.Errorf("Domain with UUID %s already exists", cfg.Name)
	}

	cfgXML, err := cfg.Marshal()
	if err != nil {
		return err
	}
	s.domain, err = s.hypervisor.DomainCreateXML(cfgXML,
		libvirt.DOMAIN_START_AUTODESTROY|libvirt.DOMAIN_START_VALIDATE)
	if err != nil {
		return err
	}

	defer func() {
		s.domain.Destroy()
		s.domain.Free()
	}()

	eventID, err := s.hypervisor.DomainEventLifecycleRegister(dom,
		s.domainLifecycleEvent)
	if err != nil {
		return err
	}

	// State may have changed between starting it and registering
	// for lifecycle events so check again now. After this point
	// we're guaranteed an event
	isActive, err := s.domain.IsActive()
	if err != nil {
		return err
	}
	if isActive {
		// We're running now, so block until we
		// either see the guest shutdown event,
		// or get a signal indicating we should
		// quit.
		select {
		case _ = <-s.shutdown:
			glog.V(1).Info("Saw guest shutdown, exiting")
			// nada
		case _ = <-s.sighandler:
			// We set the domain to auto-destroy
			// which means libvirtd will kill
			// it when exit, but by using Destroy()
			// we ensure libvirtd has sychronously
			// cleaned up the domain before we
			// exit. The exception of course is
			// SIG_KILL which is uncatchable - we
			// get only get async cleanup in that
			// case
			glog.V(1).Info("Got signal, killing guest")
			for {
				err := s.domain.Destroy()
				if err == nil {
					glog.V(1).Info("Killed guest, exiting")
					break
				}
				glog.V(1).Info("Error killing guest %s, retrying", err)
			}
		}
	} else {
		glog.V(1).Info("Guest already shutdown, exiting")
	}

	s.hypervisor.DomainEventDeregister(eventID)

	return nil
}
