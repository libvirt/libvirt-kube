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
	"fmt"
	"path"
	"regexp"

	"github.com/golang/glog"
	"github.com/libvirt/libvirt-go"
	"github.com/libvirt/libvirt-go-xml"

	"libvirt.org/libvirt-kube/pkg/api"
	apiv1 "libvirt.org/libvirt-kube/pkg/api/v1alpha1"
)

type RepositoryJobAction string

var RepositoryJobActionDelete RepositoryJobAction = "delete"
var RepositoryJobActionCreate RepositoryJobAction = "create"

type RepositoryJobCreate struct {
	file       *RepositoryFile
	name       string
	pool       *libvirt.StoragePool
	allocation uint64
	capacity   uint64
	format     string

	// output var
	vol *libvirt.StorageVol
}

type RepositoryJobDelete struct {
	vol *libvirt.StorageVol
}

type RepositoryJobResize struct {
	vol  *libvirt.StorageVol
	size uint64
}

type RepositoryJob interface {
	Process() error
	Finish(*Repository) error
}

type RepositoryFile struct {
	resource *apiv1.Virtimagefile
	vol      *libvirt.StorageVol
}

type Repository struct {
	repoclient *api.VirtimagerepoClient
	fileclient *api.VirtimagefileClient

	// API representation of resource
	resource *apiv1.Virtimagerepo

	pendingJobs   chan RepositoryJob
	completedJobs chan RepositoryJob

	// Path to storage
	path string

	poolname string

	pool *libvirt.StoragePool

	files []*RepositoryFile
}

func escapeFilename(name string) string {
	re := regexp.MustCompile("/")
	return re.ReplaceAllLiteralString(name, "_")
}

func escapeObjname(path string) string {
	re := regexp.MustCompile("[^a-zA-Z0-9_-]")
	return re.ReplaceAllLiteralString(path, "_")
}
func makeVolName(path, format string) string {
	base := escapeObjname(path)
	return fmt.Sprintf("%s.%s", base, format)
}

func (j *RepositoryJobCreate) Process() error {
	glog.V(1).Infof("Job create %s %s %d %d", j.name, j.format, j.capacity, j.allocation)

	volCFG := &libvirtxml.StorageVolume{
		Type: "file",
		Name: j.name,
		Capacity: &libvirtxml.StorageVolumeSize{
			Unit:  "bytes",
			Value: j.capacity,
		},
		Allocation: &libvirtxml.StorageVolumeSize{
			Unit:  "bytes",
			Value: j.allocation,
		},
		Target: &libvirtxml.StorageVolumeTarget{
			Format: &libvirtxml.StorageVolumeTargetFormat{
				Type: j.format,
			},
		},
	}

	volXML, err := volCFG.Marshal()
	if err != nil {
		return err
	}

	vol, err := j.pool.StorageVolCreateXML(volXML, 0)
	if err != nil {
		return err
	}

	j.vol = vol
	j.pool.Free()

	return nil
}

func (j *RepositoryJobCreate) Finish(r *Repository) error {
	glog.V(1).Infof("Finishing create %s ", j.name)
	for _, file := range r.files {
		if file == j.file {
			if file.vol == nil {
				if j.vol == nil {
					file.resource.Status.Phase = apiv1.VirtimagefileFailed
				} else {
					file.resource.Status.Phase = apiv1.VirtimagefileAvailable
					file.vol = j.vol
				}
			}
			return nil
		}
	}

	if j.vol != nil {
		j.vol.Free()
	}
	return nil
}

func (j *RepositoryJobDelete) Process() error {
	glog.V(1).Infof("Job delete %s", j.vol)

	err := j.vol.Delete(0)
	j.vol.Free()
	return err
}

func (j *RepositoryJobDelete) Finish(r *Repository) error {
	return nil
}

func jobWorker(pendingJobs chan RepositoryJob, completedJobs chan RepositoryJob) {
	glog.V(1).Info("Job worker running")
	for {
		job, more := <-pendingJobs
		if !more {
			break
		}

		err := job.Process()
		if err != nil {
			glog.V(1).Infof("Job %s failed with %s", job, err)
		}

		completedJobs <- job
	}
	glog.V(1).Info("Job worker exiting")
}

func CreateRepository(repoclient *api.VirtimagerepoClient, fileclient *api.VirtimagefileClient, resource *apiv1.Virtimagerepo, repopath string) *Repository {
	pendingJobs := make(chan RepositoryJob, 100)
	completedJobs := make(chan RepositoryJob, 100)

	if resource.Spec.JobWorkers == 0 {
		resource.Spec.JobWorkers = 3
	}

	for i := 0; i < int(resource.Spec.JobWorkers); i++ {
		go jobWorker(pendingJobs, completedJobs)
	}

	name := resource.Metadata.Name

	fullpath := path.Join(repopath, name)

	return &Repository{
		repoclient:    repoclient,
		fileclient:    fileclient,
		resource:      resource,
		path:          fullpath,
		poolname:      escapeObjname(name),
		pendingJobs:   pendingJobs,
		completedJobs: completedJobs,
	}
}

func (r *Repository) loadFileResources() error {
	files, err := r.fileclient.List()
	if err != nil {
		return err
	}

	r.files = make([]*RepositoryFile, len(files.Items))

	for i, file := range files.Items {
		glog.V(1).Infof("Loaded file resource %s (%s)", file, file.Metadata.Name)
		r.files[i] = &RepositoryFile{
			resource: file,
		}
	}

	return nil
}

func (r *Repository) disconnectHV() {
	for _, file := range r.files {
		if file.vol != nil {
			file.vol.Free()
			file.vol = nil
		}
	}

	if r.pool != nil {
		r.pool.Free()
		r.pool = nil
	}
}

func (r *Repository) createPool(conn *libvirt.Connect) (*libvirt.StoragePool, error) {
	poolCFG := libvirtxml.StoragePool{
		Type: "dir",
		Name: r.poolname,
		Target: &libvirtxml.StoragePoolTarget{
			Path: r.path,
		},
	}

	poolXML, err := poolCFG.Marshal()
	if err != nil {
		return nil, err
	}

	glog.V(1).Infof("Creating storage pool '%s' at '%s'", r.poolname, r.path)
	pool, err := conn.StoragePoolCreateXML(poolXML, libvirt.STORAGE_POOL_CREATE_WITH_BUILD)
	if err != nil {
		return nil, err
	}

	return pool, nil
}

func (r *Repository) connectHV(conn *libvirt.Connect) error {
	glog.V(1).Infof("Connected to HV")
	pool, err := conn.LookupStoragePoolByName(r.poolname)
	if err != nil {
		lverr, ok := err.(libvirt.Error)
		if ok && lverr.Code == libvirt.ERR_NO_STORAGE_POOL {
			pool, err = r.createPool(conn)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	r.pool = pool

	vols, err := pool.ListAllStorageVolumes(0)

	volNames := make(map[string]*libvirt.StorageVol)
	for i, vol := range vols {
		name, err := vol.GetName()
		if err != nil {
			continue
		}

		glog.V(1).Infof("Stash %s %v", name, vol)
		volNames[name] = &vols[i]
	}

	for _, file := range r.files {
		name := makeVolName(file.resource.Metadata.Name, r.resource.Spec.Format)

		vol, ok := volNames[name]

		if ok {
			delete(volNames, name)
			file.vol = vol
			file.resource.Status.Phase = apiv1.VirtimagefileAvailable
		} else {
			file.resource.Status.Phase = apiv1.VirtimagefilePending
			r.pool.Ref()
			job := &RepositoryJobCreate{
				file:     file,
				pool:     r.pool,
				name:     name,
				capacity: file.resource.Spec.Capacity,
				format:   r.resource.Spec.Format,
			}
			if r.resource.Spec.Preallocate {
				job.allocation = job.capacity
			}

			glog.V(1).Infof("Queueing create for %s", name)
			r.pendingJobs <- job
		}

		obj, err := r.fileclient.Update(file.resource)
		if err != nil {
			glog.V(1).Infof("Unable to update file status %s", err)
			continue
		}
		file.resource = obj
	}

	for name, vol := range volNames {
		n1, _ := vol.GetName()
		glog.V(1).Infof("Queue delete for %s %s %v", name, n1, vol)
		job := &RepositoryJobDelete{
			vol: vol,
		}
		r.pendingJobs <- job
	}

	return nil
}

func (r *Repository) refreshVol(file *RepositoryFile) error {
	info, err := file.vol.GetInfo()
	if err != nil {
		return err
	}

	file.resource.Status.Capacity = info.Capacity
	file.resource.Status.Usage = info.Allocation

	if false {
		info, err = file.vol.GetInfoFlags(libvirt.STORAGE_VOL_GET_PHYSICAL)
		if err != nil {
			return err
		}

		file.resource.Status.Length = info.Allocation
	}

	return nil
}

func (r *Repository) refreshHV() error {
	glog.V(1).Infof("Refresh HV status")
	info, err := r.pool.GetInfo()
	if err != nil {
		return err
	}

	r.resource.Status.Capacity = info.Capacity
	r.resource.Status.Allocation = info.Allocation

	// XXX update Commitment field - hard....

	for _, file := range r.files {
		if file.vol == nil {
			glog.V(1).Infof("Vol %s not ready, skipping refresh", file.resource.Metadata.Name)
			continue
		}

		glog.V(1).Infof("Vol %s ready, refresh", file.resource.Metadata.Name)
		err = r.refreshVol(file)
		if err != nil {
			glog.Errorf("Unable to refresh vol info %s", err)
			file.vol.Free()
			file.vol = nil
			file.resource.Status.Phase = apiv1.VirtimagefileFailed
		}

		obj, err := r.fileclient.Update(file.resource)
		if err != nil {
			glog.Errorf("Unable to update image file info %s", err)
			// XXX deal with fact it might be been deleted ?
		}

		file.resource = obj
	}

	return nil
}

func (r *Repository) update(conn *libvirt.Connect) error {
	glog.V(1).Info("Updating existing record")

	for done := false; !done; {
		select {
		case job := <-r.completedJobs:
			job.Finish(r)
		default:
			done = true
		}
	}

	if conn == nil {
		if r.pool == nil {
			return nil
		}
		r.disconnectHV()
		r.resource.Status.Phase = apiv1.VirtimagerepoOffline
	} else {
		if r.pool == nil {
			err := r.connectHV(conn)
			if err != nil {
				glog.Errorf("Unable to initialize HV %s", err)
				r.resource.Status.Phase = apiv1.VirtimagerepoFailed
			} else {
				r.resource.Status.Phase = apiv1.VirtimagerepoReady
			}
		}
		if r.pool != nil {
			err := r.refreshHV()
			if err != nil {
				r.disconnectHV()
				r.resource.Status.Phase = apiv1.VirtimagerepoFailed
			}
		}
	}

	obj, err := r.repoclient.Update(r.resource)

	if err != nil {
		glog.Errorf("Unable to update image repo info %s", err)
		return err
	}

	r.resource = obj

	glog.V(1).Infof("Repo Result %s %d", obj, obj.Metadata.ResourceVersion)

	return nil
}
