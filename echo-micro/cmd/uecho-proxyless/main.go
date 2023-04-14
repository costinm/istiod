// Copyright Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/costinm/grpc-mesh/echo-micro/server"
	"google.golang.org/grpc/admin"
	xdscreds "google.golang.org/grpc/credentials/xds"
	"google.golang.org/grpc/xds"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/grpclog"
)

var log = grpclog.Component("echo")

func Run(lis net.Listener) error {
	// Hack: this is expected to be started with GRPC_ env variable for bootstrap.
	// If not found, will use the internal API to initialize.

	// Replaces: creds := insecure.NewCredentials()
	creds, _ := xdscreds.NewServerCredentials(xdscreds.ServerOptions{FallbackCreds: insecure.NewCredentials()})

	grpcOptions := []grpc.ServerOption{
		grpc.Creds(creds),
	}

	// Replaces: grpc.NewServer(grpcOptions...)
	grpcServer := xds.NewGRPCServer(grpcOptions...)

	// grpcdebug support
	admin.Register(grpcServer)

	h := &server.EchoGrpcHandler{}
	h.Register(grpcServer)

	return grpcServer.Serve(lis)
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatal(err)
	}

	err = Run(lis)
	if err != nil {
		log.Fatal(err)
	}

	// Wait for the process to be shutdown.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
