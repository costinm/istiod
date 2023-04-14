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
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/costinm/grpc-mesh/echo-micro/server"
	"github.com/costinm/grpc-mesh/gen/proto/go/proto"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	om "github.com/grpc-ecosystem/go-grpc-prometheus"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"log"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"

	"google.golang.org/grpc/admin"
	"google.golang.org/grpc/reflection"
)

var (
	// Create a metrics registry.
	reg = prometheus.NewRegistry()

	// Create some standard server metrics.
	grpcMetrics = om.NewServerMetrics()

	// Create a customized counter metric.
	customizedCounterMetric = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "demo_server_say_hello_method_handle_count",
		Help: "Total number of RPCs handled on the server.",
	}, []string{"name"})
)

func init() {
	reg.MustRegister(grpcMetrics, customizedCounterMetric)
	customizedCounterMetric.WithLabelValues("Test")

	// Similar with pilot-agent
	//wrapped := prometheus.WrapRegistererWithPrefix("uecho_", prometheus.Registerer(reg))
	wrapped := reg

	wrapped.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	wrapped.MustRegister(collectors.NewGoCollector())

	http.Handle("/metrics/", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
}

func Run(lis net.Listener) error {
	h := &server.EchoGrpcHandler{}

	creds := insecure.NewCredentials()

	grpcOptions := []grpc.ServerOption{
		grpc.Creds(creds),
		grpc.StreamInterceptor(grpcMetrics.StreamServerInterceptor()),
		grpc.UnaryInterceptor(grpcMetrics.UnaryServerInterceptor()),
	}

	grpcServer := grpc.NewServer(grpcOptions...)

	proto.RegisterEchoTestServiceServer(grpcServer, h)
	admin.Register(grpcServer)
	reflection.Register(grpcServer)

	go func() {
		err := grpcServer.Serve(lis)
		if err != nil {
			panic(err)
		}
	}()
	// Initialize all metrics.
	grpcMetrics.InitializeMetrics(grpcServer)

	// Status on 8081
	// rpcz, tracez
	go http.ListenAndServe("127.0.0.1:9381", http.DefaultServeMux)
	return nil
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9382"
	}
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}

	Run(lis)

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
}
