package main

import (
	"context"
	"fmt"
	cri "github.com/kubernetes/kubernetes/staging/src/k8s.io/cri-api/pkg/apis/runtime/v1alpha2"
	"google.golang.org/grpc"
	"time"
)

const (
	address = "localhost:50052"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		fmt.Printf("did not connect: %v\n", err)
		return
	}
	defer conn.Close()

	//c := cri.NewRuntimeServiceClient(conn)

	ic := cri.NewImageServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	image := &cri.ImageSpec{Image: "docker.io/alpine:latest"}

	resp, err := ic.PullImage(ctx, &cri.PullImageRequest{Image: image})
	if err != nil {
		fmt.Printf("Error making pull image request: %v\n", err)
		return
	}
	fmt.Printf("Image pulled: %v\n", resp)

	image = &cri.ImageSpec{Image: "docker.io/hello-world"}
	resp, err = ic.PullImage(ctx, &cri.PullImageRequest{Image: image})
	if err != nil {
		fmt.Printf("Error making pull image request: %v\n", err)
		return
	}
	fmt.Printf("Image pulled: %v\n", resp)
}
