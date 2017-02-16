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
	vol      *libvirt.StorageVol
	size     uint64
	allocate bool
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

	files map[string]*RepositoryFile
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

func (j *RepositoryJobResize) Process() error {
	glog.V(1).Infof("Job resize %s", j.vol)

	flags := libvirt.STORAGE_VOL_RESIZE_SHRINK
	if j.allocate {
		flags |= libvirt.STORAGE_VOL_RESIZE_ALLOCATE
	}
	err := j.vol.Resize(j.size, flags)
	j.vol.Free()
	return err
}

func (j *RepositoryJobResize) Finish(r *Repository) error {
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

func (r *Repository) volRepoMatches(file *apiv1.Virtimagefile) bool {
	if file.Spec.RepoName != r.resource.Metadata.Name {
		glog.V(1).Infof("Ignoring vol '%s' for repo '%s'", file.Metadata.Name, file.Spec.RepoName)
		return false
	}
	return true
}

func (r *Repository) loadFileResources() error {
	files, err := r.fileclient.List()
	if err != nil {
		return err
	}

	r.files = make(map[string]*RepositoryFile)

	for _, file := range files.Items {
		if !r.volRepoMatches(file) {
			continue
		}
		name := makeVolName(file.Metadata.Name, r.resource.Spec.Format)

		glog.V(1).Infof("Loaded file resource %s (%s)", file, file.Metadata.Name)
		r.files[name] = &RepositoryFile{
			resource: file,
		}
	}

	return nil
}

func (r *Repository) createFileVolume(name string) {
	file := r.files[name]
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

	obj, err := r.fileclient.Update(file.resource)
	if err != nil {
		glog.V(1).Infof("Unable to update file status %s", err)
		return
	}
	file.resource = obj
}

func (r *Repository) loadVolumes() error {
	glog.V(1).Infof("Loading storage volumes from pool")
	vols, err := r.pool.ListAllStorageVolumes(0)
	if err != nil {
		return err
	}

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

			// XXX might need to resize the vol

			obj, err := r.fileclient.Update(file.resource)
			if err != nil {
				glog.V(1).Infof("Unable to update file status %s", err)
				continue
			}
			file.resource = obj
		} else {
			r.createFileVolume(name)
		}
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

func (r *Repository) refreshSizes() error {
	glog.V(1).Infof("Refresh HV status")
	info, err := r.pool.GetInfo()
	if err != nil {
		return err
	}

	r.resource.Status.Capacity = info.Capacity
	r.resource.Status.Allocation = info.Allocation

	// XXX update Commitment field - hard....

	for name, file := range r.files {
		if file.vol == nil {
			glog.V(1).Infof("Vol %s not ready, skipping refresh", name)
			continue
		}

		glog.V(1).Infof("Vol %s ready, refresh", name)
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

func (r *Repository) saveRepo() error {
	obj, err := r.repoclient.Update(r.resource)

	if err != nil {
		glog.Errorf("Unable to update image repo info %s", err)
		return err
	}

	glog.V(1).Infof("Repo origianalk %s %d", r.resource, r.resource.Metadata.ResourceVersion)
	r.resource = obj

	glog.V(1).Infof("Repo Result %s %d", obj, obj.Metadata.ResourceVersion)
	return nil
}

func (r *Repository) Refresh() error {
	glog.V(1).Info("Updating repo state")

	for done := false; !done; {
		select {
		case job := <-r.completedJobs:
			job.Finish(r)
		default:
			done = true
		}
	}

	if r.pool == nil {
		return nil
	}

	err := r.refreshSizes()
	if err != nil {
		glog.V(1).Infof("Failed refreshing sizes %s", err)
		r.resource.Status.Phase = apiv1.VirtimagerepoFailed
	} else {
		r.resource.Status.Phase = apiv1.VirtimagerepoReady
	}

	if err = r.saveRepo(); err != nil {
		return err
	}

	return nil
}

func (r *Repository) SetPool(pool *libvirt.StoragePool) error {
	glog.V(1).Infof("Setting pool %v", pool)
	r.pool = pool
	r.pool.Ref()

	err := r.loadVolumes()
	if err != nil {
		glog.V(1).Infof("Failed loading volumes %s", err)
		r.resource.Status.Phase = apiv1.VirtimagerepoFailed
	} else {
		r.resource.Status.Phase = apiv1.VirtimagerepoReady
	}

	if err := r.saveRepo(); err != nil {
		return err
	}
	return nil
}

func (r *Repository) UnsetPool() error {
	glog.V(1).Infof("Unsetting pool %v", r.pool)
	r.pool.Free()
	r.pool = nil

	for _, file := range r.files {
		if file.vol != nil {
			file.vol.Free()
			file.vol = nil
		}
	}

	r.resource.Status.Phase = apiv1.VirtimagerepoOffline

	if err := r.saveRepo(); err != nil {
		return err
	}
	return nil
}

func (r *Repository) AddFile(file *apiv1.Virtimagefile) {
	if !r.volRepoMatches(file) {
		return
	}

	name := makeVolName(file.Metadata.Name, r.resource.Spec.Format)

	_, ok := r.files[name]

	if ok {
		return
	}

	r.files[name] = &RepositoryFile{
		resource: file,
	}

	r.createFileVolume(name)
}

func (r *Repository) ModifyFile(file *apiv1.Virtimagefile) {
	if !r.volRepoMatches(file) {
		return
	}

	name := makeVolName(file.Metadata.Name, r.resource.Spec.Format)

	fileState, ok := r.files[name]

	if !ok {
		return
	}

	if file.Metadata.ResourceVersion == fileState.resource.Metadata.ResourceVersion {
		glog.V(1).Infof("Version did not change, ignoring modify")
		return
	}

	if file.Spec.Capacity != fileState.resource.Spec.Capacity {
		glog.V(1).Infof("Queue resize for %s %v", name, fileState.vol)
		fileState.vol.Ref()
		job := &RepositoryJobResize{
			vol:  fileState.vol,
			size: file.Spec.Capacity,
		}
		if r.resource.Spec.Preallocate {
			job.allocate = true
		}
		r.pendingJobs <- job
	}

	fileState.resource = file
}

func (r *Repository) DeleteFile(file *apiv1.Virtimagefile) {
	if !r.volRepoMatches(file) {
		return
	}

	name := makeVolName(file.Metadata.Name, r.resource.Spec.Format)

	fileState, ok := r.files[name]

	if !ok {
		return
	}

	if fileState.vol != nil {
		glog.V(1).Infof("Queue delete for %s %v", name, fileState.vol)
		job := &RepositoryJobDelete{
			vol: fileState.vol,
		}
		r.pendingJobs <- job
		fileState.vol = nil
	}
	delete(r.files, name)
}
