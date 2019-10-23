package main

import (
	"context"
	"fmt"
	"github.com/containers/libpod/libpod"
	"github.com/containers/libpod/libpod/image"
	ns "github.com/containers/libpod/pkg/namespaces"
	createconfig "github.com/containers/libpod/pkg/spec"
	"github.com/containers/libpod/pkg/util"
	"github.com/containers/storage/pkg/reexec"
	cri "github.com/kubernetes/kubernetes/staging/src/k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"google.golang.org/grpc"
	"net"
	"os"
	"syscall"
)

const (
	port = ":50052"
)

var rtime *libpod.Runtime
var libpodCtx *context.Context

type server struct {
	cri.RuntimeServiceServer
}

// Create a container with the designated image
func (s *server) CreateContainer(ctx context.Context, in *cri.CreateContainerRequest) (*cri.CreateContainerResponse, error) {
	imageName := in.Config.Image.Image
	pulledImage, err := rtime.ImageRuntime().New(*libpodCtx, imageName, "", "", os.Stdout, &image.DockerRegistryOptions{}, image.SigningOptions{}, nil, util.PullImageMissing)
	if err != nil {
		fmt.Printf("%v", err)
	}

	// Inspect the image we pulled so can create from it
	imageData, err := pulledImage.Inspect(ctx)
	if err != nil {
		fmt.Printf("%v", err)
	}

	// idmappings are required for creating a container
	idmappings, err := util.ParseIDMapping(ns.UsernsMode(""), []string{}, []string{}, "", "")
	if err != nil {
		return nil, err
	}

	cc := createconfig.CreateConfig{
		Command:    imageData.Config.Cmd,
		Detach:     true,
		IDMappings: idmappings,
		Image:      imageName,
		ImageID:    pulledImage.ID(),
		Network:    in.PodSandboxId,
		PodmanPath: "/usr/bin/podman",
		StopSignal: syscall.SIGTERM,
		WorkDir:    "/",
	}

	// Create the spec from our configuration
	containerSpec, options, err := cc.MakeContainerConfig(rtime, nil)
	if err != nil {
		return nil, err
	}

	// Create the container from the spec
	ctr, err := rtime.NewContainer(*libpodCtx, containerSpec, options...)
	if err != nil {
		return nil, err
	}

	return &cri.CreateContainerResponse{ContainerId: ctr.ID()}, nil
}

// Start the container designated by the containerid in the request
func (s *server) StartContainer(ctx context.Context, in *cri.StartContainerRequest) (*cri.StartContainerResponse, error) {
	ctn, err := rtime.GetContainer(in.ContainerId)
	if err != nil {
		return nil, err
	}

	if err = ctn.Start(*libpodCtx, false); err != nil {
		return nil, err
	}

	return &cri.StartContainerResponse{}, nil
}

func (s *server) StopContainer(ctx context.Context, in *cri.StopContainerRequest) (*cri.StopContainerResponse, error) {
	ctn, err := rtime.GetContainer(in.ContainerId)
	if err != nil {
		return nil, err
	}

	// CRI stopcontainerrequest takes a timeout param. does libpod or implement here?
	if err = ctn.Stop(); err != nil {
		return nil, err
	}

	return &cri.StopContainerResponse{}, nil
}

func main() {
	// This is required for containers storage
	if reexec.Init() {
		return
	}

	// Create a context useful for everything
	newLibpodCtx := context.TODO()
	libpodCtx = &newLibpodCtx

	// Get a libpod runtime
	runtime, err := libpod.NewRuntime(*libpodCtx)
	rtime = runtime
	if err != nil {
		fmt.Printf("Error creating libpod runtime: %v\n", err)
		return
	}

	// Listen on the designated port
	lis, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Printf("failed to listen: %v", err)
	}

	// Create a grpc runtime service server
	s := grpc.NewServer()
	cri.RegisterRuntimeServiceServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		fmt.Printf("failed to serve: %v", err)
	}
}
