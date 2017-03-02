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

// Virtimagefile defines info about a node able to run KVM guests
type Virtimagefile struct {
	v1.TypeMeta `json:",inline"`
	Metadata    v1.ObjectMeta `json:"metadata"`

	Spec   VirtimagefileSpec   `json:"spec"`
	Status VirtimagefileStatus `json:"status"`
}

// VirtimagefileList is a list of Virtimagefiles.
type VirtimagefileList struct {
	v1.TypeMeta `json:",inline"`
	Metadata    v1.ListMeta `json:"metadata"`

	Items []*Virtimagefile `json:"items"`
}

type VirtimagefileStatus struct {
	Phase VirtimagefilePhase `json:"phase"`

	// Physical usage of the file on underlying storage
	// - May be less than length if the file is sparse
	// - May be greater than length if the FS has
	//   pre-emptively reserved extra blocks for future
	//   size growth
	Usage uint64 `json:"usage"`

	// Reported length of the file on underlying storage
	Length uint64 `json:"length"`

	// Current logical capacity - may different from spec
	// capacity if a resize is pending
	Capacity uint64 `json:"capacity"`
}

type VirtimagefilePhase string

const (
	// The image file does not yet exist
	VirtimagefilePending = "Pending"
	// The image file exists
	VirtimagefileAvailable = "Available"
	// The image file failed to create
	VirtimagefileFailed = "Failed"
)

type VirtimagefileStreamAccessMode string

const (
	VirtimagefileStreamUpload   = "Upload"
	VirtimagefileStreamDownload = "Download"
	VirtimagefileStreamBoth     = "Both"
)

// VirtimagefileSpec holds specification parameters of a Virtimagefile deployment.
type VirtimagefileSpec struct {
	// Name of Virtimagerepo resource that owns this
	RepoName string `json:"repoName"`

	// Name of Virtimagefile resource that backs this
	BackingImageFile string `json:"backingImageFile"`

	AccessMode VirtimagefileAccessMode `json:"accessMode"`

	// Logical size of disk payload
	Capacity uint64 `json:"capacity"`

	Stream VirtimagefileStream `json:"stream"`
}

type VirtimagefileStream struct {
	// Name of a 'secret' object providing an access
	// control token to grant permission for upload
	// or download of the content
	TokenSecret string `json:"tokenSecret"`

	AccessMode VirtimagefileStreamAccessMode `json:"accessMode"`
}

type VirtimagefileAccessMode string

const (
	VirtimagefileReadWriteOnce VirtimagefileAccessMode = "ReadWriteOnce"
	VirtimagefileReadeOnlyMany VirtimagefileAccessMode = "ReadOnlyMany"
	VirtimagefileReadWriteMany VirtimagefileAccessMode = "ReadWriteMany"
)

// Required to satisfy Object interface
func (ni *Virtimagefile) GetObjectKind() schema.ObjectKind {
	return &ni.TypeMeta
}

// Required to satisfy ObjectMetaAccessor interface
func (ni *Virtimagefile) GetObjectMeta() v1.Object {
	return &ni.Metadata
}

// Required to satisfy Object interface
func (ni *VirtimagefileList) GetObjectKind() schema.ObjectKind {
	return &ni.TypeMeta
}

// Required to satisfy ListMetaAccessor interface
func (ni *VirtimagefileList) GetListMeta() v1.List {
	return &ni.Metadata
}
