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

package imagerepo

import (
	"reflect"
	"time"

	"github.com/golang/glog"
	"github.com/libvirt/libvirt-go"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	kubeapi "k8s.io/client-go/pkg/api"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"libvirt.org/libvirt-kube/pkg/api"
	apiv1 "libvirt.org/libvirt-kube/pkg/api/v1alpha1"
	"libvirt.org/libvirt-kube/pkg/libvirtutil"
)

type Service struct {
	poolManager *PoolManager
	fileMonitor watch.Interface
	conn        *libvirt.Connect
	connNotify  chan libvirtutil.ConnectEvent
	clientset   *kubernetes.Clientset
	repo        *Repository
}

func getKubeConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func NewService(libvirtURI string, kubeconfigfile string, reponame string, repopath string) (*Service, error) {
	kubeconfig, err := getKubeConfig(kubeconfigfile)
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	err = api.RegisterVirtimagerepo(clientset)
	if err != nil {
		return nil, err
	}

	err = api.RegisterVirtimagefile(clientset)
	if err != nil {
		return nil, err
	}

	imagerepoclient, err := api.NewVirtimagerepoClient(kubeapi.NamespaceDefault, kubeconfig)
	if err != nil {
		return nil, err
	}

	imagefileclient, err := api.NewVirtimagefileClient(kubeapi.NamespaceDefault, kubeconfig)
	if err != nil {
		return nil, err
	}

	imagerepo, err := imagerepoclient.Get(reponame)
	if err != nil {
		return nil, err
	}

	fileMonitor, err := imagefileclient.Watch()
	if err != nil {
		return nil, err
	}

	glog.V(1).Infof("Got repo %s", imagerepo)

	svc := &Service{
		poolManager: NewPoolManager(reponame, repopath),
		fileMonitor: fileMonitor,
		connNotify:  make(chan libvirtutil.ConnectEvent, 1),
		clientset:   clientset,
		repo:        CreateRepository(imagerepoclient, imagefileclient, imagerepo, repopath),
	}

	libvirtutil.OpenConnect(libvirtURI, svc.connNotify)

	return svc, nil
}

func (s *Service) connectReady(conn *libvirt.Connect) {
	glog.V(1).Info("Got connection ready event")
	s.conn = conn

	s.poolManager.Load(conn)

	conn.Close()
}

func (s *Service) connectFailed() {
	glog.V(1).Info("Got connection failed event")
	s.repo.UnsetPool()
	s.repo.Refresh()
	s.conn.Close()
	s.conn = nil
}

func (s *Service) Run() error {
	glog.V(1).Info("Running image repo service")

	ticker := time.NewTicker(time.Second * 15)

	err := s.repo.loadFileResources()
	if err != nil {
		return err
	}

	for {
		select {
		case hypEvent := <-s.connNotify:
			switch hypEvent.Type {
			case libvirtutil.ConnectReady:
				s.connectReady(hypEvent.Conn)
			case libvirtutil.ConnectFailed:
				s.connectFailed()
			}
		case pool := <-s.poolManager.Notify:
			glog.V(1).Infof("Got pool ready %v", pool)
			// Connection might have closed in meanwhile so check
			if s.conn != nil {
				err := s.repo.SetPool(pool)
				if err != nil {
					s.repo.Refresh()
				}
			}
			pool.Free()
		case objEvent := <-s.fileMonitor.ResultChan():
			if objEvent.Type == watch.Error {
				glog.V(1).Infof("Got file error %s", objEvent.Object)
				continue
			}
			glog.V(1).Infof("Object %s %s", objEvent.Type, objEvent.Object)

			imagefile, ok := objEvent.Object.(*apiv1.Virtimagefile)
			if !ok {
				glog.V(1).Infof("Object wasn't virtimagefile %s", objEvent.Object, reflect.TypeOf(objEvent.Object))
				continue
			}
			glog.V(1).Infof("Object %s %s", objEvent.Type, imagefile.Metadata.Name)
			switch objEvent.Type {
			case watch.Added:
				s.repo.AddFile(imagefile)
			case watch.Modified:
				s.repo.ModifyFile(imagefile)
			case watch.Deleted:
				s.repo.DeleteFile(imagefile)
			}
		case <-ticker.C:
			glog.V(1).Info("Updating repo")
			s.repo.Refresh()
		}
	}

	return nil
}
