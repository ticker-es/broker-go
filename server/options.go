package server

import "google.golang.org/grpc"

func ListenAddress(address string) Option {
	return func(s *Server) {
		s.listen = address
	}
}

func grpcWrapper(opts ...grpc.ServerOption) Option {
	return func(s *Server) {
		s.grpcServerOptions = append(s.grpcServerOptions, opts...)
	}
}
