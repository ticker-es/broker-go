package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"os"

	"google.golang.org/grpc/credentials"

	. "github.com/mtrense/soil/config"
	"github.com/mtrense/soil/logging"
	log "github.com/mtrense/soil/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ticker-es/broker-go/eventstore/memory"
	"github.com/ticker-es/broker-go/server"
)

var (
	version = "none"
	commit  = "none"
	app     = NewCommandline("ticker-broker",
		Short("Ticker Broker"),
		FlagLogLevel("warn"),
		FlagLogFormat(),
		FlagLogFile(),
		SubCommand("server",
			Short("Run the ticker server"),
			Flag("listen", Str(":6677"), Description("Address to listen for grpc connections"), Mandatory(), Persistent(), Env()),
			Flag("database", Str("localhost:5432"), Description("Database server to connect to"), Mandatory(), Persistent(), Env()),
			Flag("insecure", Bool(), Description("Run in insecure mode. Not recommended in production"), Persistent(), Env()),
			Flag("tls-key", Str("tls/server.key"), Description("TLS key to use"), Persistent(), EnvName("tls_key")),
			Flag("tls-cert", Str("tls/server.crt"), Description("TLS certificate to use"), Persistent(), EnvName("tls_cert")),
			Flag("client-ca", Str("tls/ca.crt"), Description("CA to verify client certs against"), Persistent(), EnvName("client_ca")),
			Run(executeServer),
		),
		Version(version, commit),
		Completion(),
	).GenerateCobra()
)

func init() {
	EnvironmentConfig("TICKER")
	log.ConfigureDefaultLogging()
}

func main() {
	if err := app.Execute(); err != nil {
		panic(err)
	}
}

func executeServer(cmd *cobra.Command, args []string) {
	listen := viper.GetString("listen")
	stream := memory.NewMemoryEventStream(memory.NewMemorySequenceStore())
	tlsConfig := &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    readCACerts(viper.GetString("client_ca")),
		Certificates: readServerCert(),
	}
	srv := server.NewServer(version, stream, credentials.NewTLS(tlsConfig), server.ListenAddress(listen))
	log.L().Info().Str("listen-addr", listen).Int("pid", os.Getpid()).Msg("Server starting")
	if err := srv.Start(); err != nil {
		panic(err)
	}
}

func readCACerts(caCertFiles ...string) *x509.CertPool {
	caCerts := x509.NewCertPool()
	for _, caCertFile := range caCertFiles {
		if caCertData, err := ioutil.ReadFile(caCertFile); err == nil {
			if !caCerts.AppendCertsFromPEM(caCertData) {
				logging.L().Error().Str("filename", caCertFile).Msg("Could not parse CA Certificate from PEM")
			}
		} else {
			logging.L().Err(err).Msg("Could not read CA Certificate")
		}
	}
	return caCerts
}

func readServerCert() []tls.Certificate {
	var certificates []tls.Certificate
	if cert, err := tls.LoadX509KeyPair(viper.GetString("tls_cert"), viper.GetString("tls_key")); err == nil {
		certificates = append(certificates, cert)
	} else {
		logging.L().Err(err).Msg("Could not read server certificate/key")
	}
	return certificates
}
