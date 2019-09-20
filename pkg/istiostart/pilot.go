package istiostart


//func StartPilotOld(baseDir string, mcfg *meshv1.MeshConfig) error {
//	stop := make(chan struct{})
//
//	// Create a test pilot discovery service configured to watch the tempDir.
//	args := bootstrap.PilotArgs{
//		Namespace: "testing",
//		DiscoveryOptions: envoy.DiscoveryServiceOptions{
//			HTTPAddr:        ":12007",
//			GrpcAddr:        ":12010",
//			SecureGrpcAddr:  ":12011",
//			EnableCaching:   true,
//			EnableProfiling: true,
//		},
//
//		Mesh: bootstrap.MeshArgs{
//
//			MixerAddress:    "localhost:9091",
//			RdsRefreshDelay: types.DurationProto(10 * time.Millisecond),
//		},
//		Config: bootstrap.ConfigArgs{},
//		Service: bootstrap.ServiceArgs{
//			// Using the Mock service registry, which provides the hello and world services.
//			Registries: []string{
//				string(serviceregistry.MCPRegistry)},
//		},
//
//		// MCP is messing up with the grpc settings...
//		MCPMaxMessageSize:        1024 * 1024 * 64,
//		MCPInitialWindowSize:     1024 * 1024 * 64,
//		MCPInitialConnWindowSize: 1024 * 1024 * 64,
//
//		MeshConfig:       mcfg,
//		KeepaliveOptions: keepalive.DefaultOption(),
//	}
//
//	// Create and setup the controller.
//	s, err := bootstrap.NewServer(args)
//	if err != nil {
//		return err
//	}
//
//	// Start the server.
//	if err := s.Start(stop); err != nil {
//		return err
//	}
//	return nil
//}
