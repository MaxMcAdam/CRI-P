package main

import (
	"context"
	"fmt"
	cri "github.com/kubernetes/kubernetes/staging/src/k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"google.golang.org/grpc"
	"os"
	"time"
)

const (
	address = "localhost:50052"
	IMAGE   = "quay.io/libpod/alpine_nginx"
)

func main() {
	image := IMAGE
	if len(os.Args) > 1 {
		image = os.Args[1]
	}
	// Create grpc connection
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("did not connect: %v\n", err)
		return
	}
	defer conn.Close()

	// Create cri runtime service client
	c := cri.NewRuntimeServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	ctnrCreateRequest := cri.CreateContainerRequest{
		Config: &cri.ContainerConfig{
			Envs:  []*cri.KeyValue{&cri.KeyValue{Key: "HZN_ORG_ID", Value: "major-peacock-icp-cluster"}},
			Image: &cri.ImageSpec{Image: image},
		},
	}

	// Create a container from the given image
	createContainerResp, err := c.CreateContainer(ctx, &ctnrCreateRequest)
	if err != nil {
		fmt.Printf("Error creating container: %v\n", err)
		return
	}
	fmt.Printf("Contianer created %v\n", createContainerResp)

	// Start the created container
	startContainerResp, err := c.StartContainer(ctx, &cri.StartContainerRequest{ContainerId: createContainerResp.ContainerId})
	if err != nil {
		fmt.Printf("Error starting container: %v\n", err)
		return
	}
	fmt.Printf("Container started %v\n", startContainerResp)

	// Stop the started container

	stopContainerResp, err := c.StopContainer(ctx, &cri.StopContainerRequest{ContainerId: createContainerResp.ContainerId})
	if err != nil {
		fmt.Printf("Error stopping container: %v\n", err)
		return
	}
	fmt.Printf("Container stopped %v\n", stopContainerResp)
}
