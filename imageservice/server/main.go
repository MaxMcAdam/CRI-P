package main

import (
	"context"
	"fmt"
	"github.com/containers/libpod/libpod"
	"github.com/containers/libpod/libpod/image"
	"github.com/containers/libpod/pkg/util"
	"github.com/containers/storage/pkg/reexec"
	cri "github.com/kubernetes/kubernetes/staging/src/k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"google.golang.org/grpc"
	"net"
	"os"
)

const (
	port = ":50052"
)

var rtime *libpod.Runtime

type imageServer struct {
	cri.ImageServiceServer
}

func (s *imageServer) PullImage(ctx context.Context, in *cri.PullImageRequest) (*cri.PullImageResponse, error) {
	pulledImage, err := rtime.ImageRuntime().New(ctx, in.Image.Image, "", "", os.Stdout, &image.DockerRegistryOptions{}, image.SigningOptions{}, nil, util.PullImageMissing)
	if err != nil {
		return nil, fmt.Errorf("Error pulling image: %v", err)
	}
	return &cri.PullImageResponse{ImageRef: pulledImage.InputName}, nil
}

func main() {
	fmt.Printf("started main\n")
	// This is required for containers storage
	if reexec.Init() {
		return
	}
	// Create a context useful for everything
	ctx := context.TODO()
	fmt.Printf("context created\n")

	// Step 1. Get a libpod runtime.  This is the entry way to using the API
	runtime, err := libpod.NewRuntime(ctx)
	if err != nil {
		fmt.Printf("Error creating libpod runtime: %v\n", err)
		return
	}
	fmt.Printf("Runtime created\n")

	rtime = runtime

	lis, err := net.Listen("tcp", port)
	if err != nil {
		fmt.Printf("failed to listen: %v", err)
	}
	fmt.Printf("Listening on %v\n", port)
	s := grpc.NewServer()

	cri.RegisterImageServiceServer(s, &imageServer{})
	if err := s.Serve(lis); err != nil {
		fmt.Printf("failed to serve: %v", err)
	}
}
