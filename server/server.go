package server

import (
	"context"
	"net"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"google.golang.org/grpc/credentials"

	"github.com/ticker-es/client-go/eventstream/base"

	"google.golang.org/grpc/stats"

	"google.golang.org/grpc/peer"

	"github.com/mtrense/soil/logging"
	"google.golang.org/grpc/reflection"

	"github.com/ticker-es/client-go/rpc"
	"google.golang.org/grpc"
)

type Server struct {
	listen            string
	grpcServerOptions []grpc.ServerOption
	version           string
	streamBackend     base.EventStream
	streamServer      *eventStreamServer
	maintenanceServer *maintenanceServer
	connectionCount   int32
	startTime         time.Time
	stopContext       context.Context
	gracefulStop      func()
	grpcServer        *grpc.Server
}

type Option = func(s *Server)

func NewServer(version string, backend base.EventStream, opts ...Option) *Server {
	srv := &Server{
		version:       version,
		streamBackend: backend,
	}
	for _, opt := range opts {
		opt(srv)
	}
	srv.streamServer = &eventStreamServer{
		server: srv,
	}
	srv.maintenanceServer = &maintenanceServer{
		server: srv,
	}
	return srv
}

func (s *Server) Start() error {
	s.startTime = time.Now()
	s.stopContext, s.gracefulStop = context.WithCancel(context.Background())
	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM, os.Kill)
	listener, err := net.Listen("tcp", ":6677")
	if err != nil {
		return err
	}
	serverOpts := []grpc.ServerOption{
		grpc.StatsHandler(s),
	}
	serverOpts = append(serverOpts, s.grpcServerOptions...)
	s.grpcServer = grpc.NewServer(serverOpts...)
	go func() {
		sig := <-signals
		switch sig {
		case os.Kill:
			s.grpcServer.Stop()
		case os.Interrupt, syscall.SIGTERM:
			s.GracefulStop()
		}
	}()
	rpc.RegisterEventStreamServer(s.grpcServer, s.streamServer)
	rpc.RegisterMaintenanceServer(s.grpcServer, s.maintenanceServer)
	reflection.Register(s.grpcServer)
	return s.grpcServer.Serve(listener)
}

func (s *Server) GracefulStop() {
	s.gracefulStop()
	s.grpcServer.GracefulStop()
}

func (s *Server) TagRPC(ctx context.Context, i *stats.RPCTagInfo) context.Context {

	return ctx
}

func (s *Server) HandleRPC(ctx context.Context, st stats.RPCStats) {
	l := logging.L().Debug()
	if p, ok := peer.FromContext(ctx); ok {
		l.Str("clientAddr", p.Addr.String())
		switch ai := p.AuthInfo.(type) {
		case credentials.TLSInfo:
			l.Bool("authenticated", true)
			l.Str("subject", ai.State.PeerCertificates[0].Subject.CommonName)
		default:
			l.Bool("authenticated", false)
		}
	}
	switch s := st.(type) {
	case *stats.Begin:
		l.Msg("RPC Call started")
	case *stats.InHeader:
		l.Str("method", s.FullMethod).Msg("RPC Call executing")
	case *stats.InPayload:
	case *stats.InTrailer:
	case *stats.OutHeader:
	case *stats.OutPayload:
	case *stats.OutTrailer:
	case *stats.End:
		l.Msg("RPC Call ended")
	}
}

func (s *Server) TagConn(ctx context.Context, i *stats.ConnTagInfo) context.Context {
	return ctx
}

func (s *Server) HandleConn(ctx context.Context, st stats.ConnStats) {
	l := logging.L().Info()
	switch st.(type) {
	case *stats.ConnBegin:
		atomic.AddInt32(&s.connectionCount, 1)
		l.Msg("Client connected")
	case *stats.ConnEnd:
		atomic.AddInt32(&s.connectionCount, -1)
		l.Msg("Client disconnected")
	}
}

func (s *Server) withGlobalStop(c context.Context) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(c)

	go func() {
		select {
		case <-ctx.Done():
		case <-s.stopContext.Done():
			cancel()
		}
	}()

	return ctx, cancel
}
