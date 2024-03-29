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

	// Create cri image service client
	ic := cri.NewImageServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	// Pull the specified image
	resp, err := ic.PullImage(ctx, &cri.PullImageRequest{Image: &cri.ImageSpec{Image: image}})
	if err != nil {
		fmt.Printf("Error making pull image request: %v\n", err)
		return
	}
	fmt.Printf("Image pulled: %v\n", resp)
}
