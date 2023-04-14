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
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/costinm/grpc-mesh/echo-micro/server"
	//grpc_zap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap"
	grpc_zap "github.com/costinm/grpc-mesh/telemetry/logs/zap"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func Run(port string) error {
	zl, _ := zap.NewDevelopment(zap.AddCallerSkip(4))
	// TODO: use zap.observer, we can use the collected info for assertions.
	grpc_zap.ReplaceGrpcLoggerV2WithVerbosity(zl, 99)

	alwaysLoggingDeciderServer := func(ctx context.Context, fullMethodName string, servingObject interface{}) bool { return true }
	alwaysLoggingDeciderClient := func(ctx context.Context, fullMethodName string) bool { return true }

	h := &server.EchoGrpcHandler{
		DialOptions: []grpc.DialOption{
			grpc.WithUnaryInterceptor(grpc_zap.UnaryClientInterceptor(zl)),
			grpc.WithStreamInterceptor(grpc_zap.StreamClientInterceptor(zl)),
			grpc.WithUnaryInterceptor(grpc_zap.PayloadUnaryClientInterceptor(zl, alwaysLoggingDeciderClient)),
			grpc.WithStreamInterceptor(grpc_zap.PayloadStreamClientInterceptor(zl, alwaysLoggingDeciderClient)),
		},
	}
	lis, err := net.Listen("tcp", port)
	if err != nil {
		return err
	}

	creds := insecure.NewCredentials()

	grpcOptions := []grpc.ServerOption{
		grpc.Creds(creds),
		grpc.StreamInterceptor(grpc_zap.StreamServerInterceptor(zl)),
		grpc.StreamInterceptor(grpc_zap.PayloadStreamServerInterceptor(zl, alwaysLoggingDeciderServer)),
		grpc.UnaryInterceptor(grpc_zap.UnaryServerInterceptor(zl)),
		grpc.UnaryInterceptor(grpc_zap.PayloadUnaryServerInterceptor(zl, alwaysLoggingDeciderServer)),
	}

	grpcServer := grpc.NewServer(grpcOptions...)
	h.Register(grpcServer)

	go func() {
		err = grpcServer.Serve(lis)
		if err != nil {
			panic(err)
		}
	}()

	return nil
}

func main() {
	// Istio already has a dependency on Zap and istio-ecosystem - set the logger
	// so we can explicitly enable level/verbosity. Alternative is to use env variables.
	// '4' skips the adapters, shows the actual code logging the info.
	// 3  works for most core - but fails for 'prefixLogger' - used by XDS

	err := Run(":9480")
	if err != nil {
		fmt.Println("Error ", err)
	}

	// Wait for the process to be shutdown.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs

}
