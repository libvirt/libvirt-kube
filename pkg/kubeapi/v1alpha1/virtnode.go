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

package v1alpha1

import (
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Virtnode defines info about a node able to run KVM guests
type Virtnode struct {
	v1.TypeMeta `json:",inline"`
	Metadata    v1.ObjectMeta `json:"metadata"`

	Spec VirtnodeSpec `json:"spec"`
}

// VirtnodeList is a list of Virtnodes.
type VirtnodeList struct {
	v1.TypeMeta `json:",inline"`
	Metadata    v1.ListMeta `json:"metadata"`

	Items []*Virtnode `json:"items"`
}

// VirtnodeSpec holds specification parameters of a Virtnode deployment.
type VirtnodeSpec struct {
	UUID      string            `json:"uuid"`
	Arch      string            `json:"arch"`
	Guests    []VirtnodeGuest   `json:"guests"`
	Resources VirtnodeResources `json:"resources"`
}

type VirtnodeGuest struct {
	Hypervisor string `json:"hypervisor"`
	Arch       string `json:"arch"`
	Type       string `json:"type"`

	Machines []string `json:"machines"`
}

type VirtnodeMemory struct {
	PageSize int    `json:"pagesize"`
	Present  uint64 `json:"present"`
	Used     uint64 `json:"used"`
}

type VirtnodeCPU struct {
	Avail int `json:"avail"`
	Used  int `json:"used"`
}

type VirtnodeNUMACell struct {
	CPU    VirtnodeCPU      `json:"cpu"`
	Memory []VirtnodeMemory `json:"memory"`
}

type VirtnodeResources struct {
	NUMACells []VirtnodeNUMACell `json:"cells"`
}

// Required to satisfy Object interface
func (ni *Virtnode) GetObjectKind() schema.ObjectKind {
	return &ni.TypeMeta
}

// Required to satisfy ObjectMetaAccessor interface
func (ni *Virtnode) GetObjectMeta() v1.Object {
	return &ni.Metadata
}

// Required to satisfy Object interface
func (ni *VirtnodeList) GetObjectKind() schema.ObjectKind {
	return &ni.TypeMeta
}

// Required to satisfy ListMetaAccessor interface
func (ni *VirtnodeList) GetListMeta() v1.List {
	return &ni.Metadata
}
