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
	"path"

	"github.com/golang/glog"
	"github.com/libvirt/libvirt-go"
	"github.com/libvirt/libvirt-go-xml"

	"libvirt.org/libvirt-kube/pkg/libvirtutil"
)

type PoolManager struct {
	Name   string
	Path   string
	Notify chan *libvirt.StoragePool
}

func NewPoolManager(name, path string) *PoolManager {
	return &PoolManager{
		Name:   name,
		Path:   path,
		Notify: make(chan *libvirt.StoragePool),
	}
}

func (m *PoolManager) create(conn *libvirt.Connect) (*libvirt.StoragePool, error) {
	name := libvirtutil.EscapeObjectName(m.Name)
	path := path.Join(m.Path, name)

	poolCFG := libvirtxml.StoragePool{
		Type: "dir",
		Name: name,
		Target: &libvirtxml.StoragePoolTarget{
			Path: path,
		},
	}

	poolXML, err := poolCFG.Marshal()
	if err != nil {
		return nil, err
	}

	glog.V(1).Infof("Creating storage pool '%s' at '%s'", name, path)
	pool, err := conn.StoragePoolCreateXML(poolXML, libvirt.STORAGE_POOL_CREATE_WITH_BUILD)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func (m *PoolManager) load(conn *libvirt.Connect) {
	glog.V(1).Infof("Loading storage pool")
	pool, err := conn.LookupStoragePoolByName(m.Name)
	if err != nil {
		lverr, ok := err.(libvirt.Error)
		if ok && lverr.Code == libvirt.ERR_NO_STORAGE_POOL {
			pool, err = m.create(conn)
			if err != nil {
				glog.V(1).Infof("Failed to create pool %s", err)
			}
		} else {
			glog.V(1).Infof("Failed to fetch pool %s", err)
		}
	}

	conn.Close()

	// pool may be nil if we had an error
	glog.V(1).Infof("Notify storage pool %v", pool)
	m.Notify <- pool
}

func (m *PoolManager) Load(conn *libvirt.Connect) {
	conn.Ref()
	go m.load(conn)
}
