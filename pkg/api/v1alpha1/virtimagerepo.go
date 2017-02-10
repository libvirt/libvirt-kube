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

// Virtimagerepo defines a Virtimagerepo deployment.
type Virtimagerepo struct {
	v1.TypeMeta `json:",inline"`
	Metadata    v1.ObjectMeta `json:"metadata"`

	Spec   VirtimagerepoSpec   `json:"spec"`
	Status VirtimagerepoStatus `json:"status"`
}

// VirtimagerepoList is a list of Virtimagerepos.
type VirtimagerepoList struct {
	v1.TypeMeta `json:",inline"`
	Metadata    v1.ListMeta `json:"metadata"`

	Items []*Virtimagerepo `json:"items"`
}

type VirtimagerepoPhase string

const (
	VirtimagerepoReady   = "Ready"
	VirtimagerepoFailed  = "Failed"
	VirtimagerepoOffline = "Offline"
)

type VirtimagerepoStatus struct {
	Phase VirtimagerepoPhase `json:"phase"`
	// Physical size of the underlying filesystem
	Capacity uint64 `json:"capacity"`
	// Total size currently allocated to images
	Allocation uint64 `json:"allocation"`
	// Total size that is committed to serving images
	// ie if all sparse images grew to their max permitted
	// size this is what would be consumed
	Commitment uint64 `json:"committment"`
}

// VirtimagerepoSpec holds specification parameters of a Virtimagerepo deployment.
type VirtimagerepoSpec struct {
	// Name of a PesistentVolumeClaim in the same namespace as the Virtimagerepo
	ClaimName string `json:"claimName"`

	// Force image format type
	Format string `json:"format"`

	Preallocate bool `json:"preallocate"`

	JobWorkers uint8 `json:"jobWorkers"`
}

// Required to satisfy Object interface
func (ni *Virtimagerepo) GetObjectKind() schema.ObjectKind {
	return &ni.TypeMeta
}

// Required to satisfy ObjectMetaAccessor interface
func (ni *Virtimagerepo) GetObjectMeta() v1.Object {
	return &ni.Metadata
}

// Required to satisfy Object interface
func (ni *VirtimagerepoList) GetObjectKind() schema.ObjectKind {
	return &ni.TypeMeta
}

// Required to satisfy ListMetaAccessor interface
func (ni *VirtimagerepoList) GetListMeta() v1.List {
	return &ni.Metadata
}
