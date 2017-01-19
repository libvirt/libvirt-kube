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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/pkg/api/v1"
)

// VirtNodeInfo defines info about a node able to run KVM guests
type VirtNodeInfo struct {
	metav1.TypeMeta `json:",inline"`
	v1.ObjectMeta   `json:"metadata,omitempty"`
	Spec            VirtNodeInfoSpec `json:"spec"`
}

// VirtNodeInfoList is a list of VirtNodeInfos.
type VirtNodeInfoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []*VirtNodeInfo `json:"items"`
}

// VirtNodeSpec holds specification parameters of a VirtNodeInfo deployment.
type VirtNodeInfoSpec struct {
	UUID      string                `json:"uuid"`
	Arch      string                `json:"arch"`
	Guests    []VirtNodeInfoGuest   `json:"guests"`
	Resources VirtNodeInfoResources `json:"resources"`
}

type VirtNodeInfoGuest struct {
	Hypervisor string `json:"hypervisor"`
	Arch       string `json:"arch"`
	Type       string `json:"type"`

	Machines []string `json:"machines"`
}

type VirtNodeInfoMemory struct {
	PageSize int `json:"pagesize"`
	Present  int `json:"present"`
	Used     int `json:"used"`
}

type VirtNodeInfoCPU struct {
	Avail int `json:"avail"`
	Used  int `json:"used"`
}

type VirtNodeInfoNUMACell struct {
	CPU    VirtNodeInfoCPU      `json:"cpu"`
	Memory []VirtNodeInfoMemory `json:"memory"`
}

type VirtNodeInfoResources struct {
	NUMACells []VirtNodeInfoNUMACell `json:"cells"`
}
