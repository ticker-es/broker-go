package main

import (
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"
	"os"

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
	stream := memory.NewMemoryEventStream(memory.NewMemoryEventStore(), memory.NewMemorySequenceStore())
	cert, err := readServerCert()
	if err != nil {
		panic(err)
	}
	srv := server.NewServer(version, stream,
		server.ListenAddress(listen),
		//server.Credentials(credentials.NewTLS(tlsConfig)),
		server.MutualTLS(cert, readCACerts()),
	)
	log.L().Info().Str("listen-addr", listen).Int("pid", os.Getpid()).Msg("Server starting")
	if err := srv.Start(); err != nil {
		panic(err)
	}
}

func readCACerts() *x509.CertPool {
	caCerts := x509.NewCertPool()
	for _, caCertFile := range viper.GetStringSlice("client_ca") {
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

func readServerCert() (tls.Certificate, error) {
	return tls.LoadX509KeyPair(viper.GetString("tls_cert"), viper.GetString("tls_key"))
}
