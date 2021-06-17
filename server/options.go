package server

import (
	"crypto/tls"
	"crypto/x509"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

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

//func Insecure() Option {
//
//}

func Credentials(tc credentials.TransportCredentials) Option {
	return grpcWrapper(grpc.Creds(tc))
}

func MutualTLS(serverCertificate tls.Certificate, clientCA *x509.CertPool) Option {
	tlsConfig := &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    clientCA,
		Certificates: []tls.Certificate{serverCertificate},
	}
	return grpcWrapper(grpc.Creds(credentials.NewTLS(tlsConfig)))
}
