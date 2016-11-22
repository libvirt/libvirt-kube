package service

import (
	"errors"
	"net"
	"os"
	"syscall"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	"k8s.io/kubernetes/pkg/kubelet/api/v1alpha1/runtime"
)

type LibvirtKubeletService struct {
	server *grpc.Server
	addr   string
}

func New(addr string) (*LibvirtKubeletService, error) {

	svc := &LibvirtKubeletService{
		server: grpc.NewServer(),
		addr:   addr,
	}

	runtime.RegisterRuntimeServiceServer(svc.server, svc)
	runtime.RegisterImageServiceServer(svc.server, svc)

	return svc, nil

}

func (s *LibvirtKubeletService) Run() error {

	err := syscall.Unlink(s.addr)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	sock, err := net.Listen("unix", s.addr)
	if err != nil {
		return err
	}

	defer sock.Close()
	return s.server.Serve(sock)
}

// Version returns the runtime name, runtime version, and runtime API version.
func (s *LibvirtKubeletService) Version(ctx context.Context, req *runtime.VersionRequest) (*runtime.VersionResponse, error) {
	return nil, errors.New("not implemented")
}

// RunPodSandbox creates and starts a pod-level sandbox. Runtimes must ensure
// the sandbox is in the ready state on success.
func (s *LibvirtKubeletService) RunPodSandbox(ctx context.Context, req *runtime.RunPodSandboxRequest) (*runtime.RunPodSandboxResponse, error) {
	return nil, errors.New("not implemented")
}

// StopPodSandbox stops any running process that is part of the sandbox and
// reclaims network resources (e.g., IP addresses) allocated to the sandbox.
// If there are any running containers in the sandbox, they must be forcibly
// terminated.
// This call is idempotent, and must not return an error if all relevant
// resources have already been reclaimed. kubelet will call StopPodSandbox
// at least once before calling RemovePodSandbox. It will also attempt to
// reclaim resources eagerly, as soon as a sandbox is not needed. Hence,
// multiple StopPodSandbox calls are expected.
func (s *LibvirtKubeletService) StopPodSandbox(ctx context.Context, req *runtime.StopPodSandboxRequest) (*runtime.StopPodSandboxResponse, error) {
	return nil, errors.New("not implemented")
}

// RemovePodSandbox removes the sandbox. If there are any running containers
// in the sandbox, they must be forcibly terminated and removed.
// This call is idempotent, and must not return an error if the sandbox has
// already been removed.
func (s *LibvirtKubeletService) RemovePodSandbox(ctx context.Context, req *runtime.RemovePodSandboxRequest) (*runtime.RemovePodSandboxResponse, error) {
	return nil, errors.New("not implemented")
}

// PodSandboxStatus returns the status of the PodSandbox.
func (s *LibvirtKubeletService) PodSandboxStatus(ctx context.Context, req *runtime.PodSandboxStatusRequest) (*runtime.PodSandboxStatusResponse, error) {
	return nil, errors.New("not implemented")
}

// ListPodSandbox returns a list of PodSandboxes.
func (s *LibvirtKubeletService) ListPodSandbox(ctx context.Context, req *runtime.ListPodSandboxRequest) (*runtime.ListPodSandboxResponse, error) {
	return nil, errors.New("not implemented")
}

// CreateContainer creates a new container in specified PodSandbox
func (s *LibvirtKubeletService) CreateContainer(ctx context.Context, req *runtime.CreateContainerRequest) (*runtime.CreateContainerResponse, error) {
	return nil, errors.New("not implemented")
}

// StartContainer starts the container.
func (s *LibvirtKubeletService) StartContainer(ctx context.Context, req *runtime.StartContainerRequest) (*runtime.StartContainerResponse, error) {
	return nil, errors.New("not implemented")
}

// StopContainer stops a running container with a grace period (i.e., timeout).
// This call is idempotent, and must not return an error if the container has
// already been stopped.
// TODO: what must the runtime do after the grace period is reached?
func (s *LibvirtKubeletService) StopContainer(ctx context.Context, req *runtime.StopContainerRequest) (*runtime.StopContainerResponse, error) {
	return nil, errors.New("not implemented")
}

// RemoveContainer removes the container. If the container is running, the
// container must be forcibly removed.
// This call is idempotent, and must not return an error if the container has
// already been removed.
func (s *LibvirtKubeletService) RemoveContainer(ctx context.Context, req *runtime.RemoveContainerRequest) (*runtime.RemoveContainerResponse, error) {
	return nil, errors.New("not implemented")
}

// ListContainers lists all containers by filters.
func (s *LibvirtKubeletService) ListContainers(ctx context.Context, req *runtime.ListContainersRequest) (*runtime.ListContainersResponse, error) {
	return nil, errors.New("not implemented")
}

// ContainerStatus returns status of the container.
func (s *LibvirtKubeletService) ContainerStatus(ctx context.Context, req *runtime.ContainerStatusRequest) (*runtime.ContainerStatusResponse, error) {
	return nil, errors.New("not implemented")
}

// ExecSync runs a command in a container synchronously.
func (s *LibvirtKubeletService) ExecSync(ctx context.Context, req *runtime.ExecSyncRequest) (*runtime.ExecSyncResponse, error) {
	return nil, errors.New("not implemented")
}

// Exec prepares a streaming endpoint to execute a command in the container.
func (s *LibvirtKubeletService) Exec(ctx context.Context, req *runtime.ExecRequest) (*runtime.ExecResponse, error) {
	return nil, errors.New("not implemented")
}

// Attach prepares a streaming endpoint to attach to a running container.
func (s *LibvirtKubeletService) Attach(ctx context.Context, req *runtime.AttachRequest) (*runtime.AttachResponse, error) {
	return nil, errors.New("not implemented")
}

// PortForward prepares a streaming endpoint to forward ports from a PodSandbox.
func (s *LibvirtKubeletService) PortForward(ctx context.Context, req *runtime.PortForwardRequest) (*runtime.PortForwardResponse, error) {
	return nil, errors.New("not implemented")
}

// UpdateRuntimeConfig updates the runtime configuration based on the given request.
func (s *LibvirtKubeletService) UpdateRuntimeConfig(ctx context.Context, req *runtime.UpdateRuntimeConfigRequest) (*runtime.UpdateRuntimeConfigResponse, error) {
	return nil, errors.New("not implemented")
}

// Status returns the status of the runtime.
func (s *LibvirtKubeletService) Status(ctx context.Context, req *runtime.StatusRequest) (*runtime.StatusResponse, error) {
	return nil, errors.New("not implemented")
}

// ListImages lists existing images.
func (s *LibvirtKubeletService) ListImages(ctx context.Context, req *runtime.ListImagesRequest) (*runtime.ListImagesResponse, error) {
	return nil, errors.New("not implemented")
}

// ImageStatus returns the status of the image. If the image is not
// present, returns nil.
func (s *LibvirtKubeletService) ImageStatus(ctx context.Context, req *runtime.ImageStatusRequest) (*runtime.ImageStatusResponse, error) {
	return nil, errors.New("not implemented")
}

// PullImage pulls an image with authentication config.
func (s *LibvirtKubeletService) PullImage(ctx context.Context, req *runtime.PullImageRequest) (*runtime.PullImageResponse, error) {
	return nil, errors.New("not implemented")
}

// RemoveImage removes the image.
// This call is idempotent, and must not return an error if the image has
// already been removed.
func (s *LibvirtKubeletService) RemoveImage(ctx context.Context, req *runtime.RemoveImageRequest) (*runtime.RemoveImageResponse, error) {
	return nil, errors.New("not implemented")
}
