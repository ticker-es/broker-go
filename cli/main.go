package main

import (
	"crypto/tls"
	"fmt"
	"github.com/ticker-es/client-go/config"
	"os"
	"strings"

	eventstream2 "github.com/ticker-es/broker-go/eventstream"

	"github.com/ticker-es/broker-go/backends"

	. "github.com/mtrense/soil/config"
	log "github.com/mtrense/soil/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
			// Server Flags
			Flag("listen", Str(":6677"), Description("Address to listen for grpc connections"), Mandatory(), Persistent(), Env()),
			Flag("insecure", Bool(), Description("Run in insecure mode. Not recommended in production"), Persistent(), Env()),
			Flag("tls-key", Str("tls/server.key"), Description("TLS key to use"), Persistent(), EnvName("tls_key")),
			Flag("tls-cert", Str("tls/server.crt"), Description("TLS certificate to use"), Persistent(), EnvName("tls_cert")),
			Flag("client-ca", Str("tls/ca.crt"), Description("CA to verify client certs against"), Persistent(), EnvName("client_ca")),
			// Backend Storage Selection
			Flag("event-store", Str("memory"), Abbr("e"), Description("Select which EventStore implementation to use"), Mandatory(), Persistent(), EnvName("event_store")),
			Flag("sequence-store", Str("memory"), Abbr("s"), Description("Select which SequenceStore implementation to use"), Mandatory(), Persistent(), EnvName("sequence_store")),
			backends.GetAllConfiguredFlags(),
			Run(executeServer),
		),
		SubCommand("list-backends",
			Short("List the available backends and their options"),
			Run(executeListBackends),
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
	eventStoreFactory := backends.LookupEventStore(viper.GetString("event_store"))
	sequenceStoreFactory := backends.LookupSequenceStore(viper.GetString("sequence_store"))
	stream := eventstream2.NewEventStream(eventStoreFactory.CreateEventStore(), sequenceStoreFactory.CreateSequenceStore())
	cert, err := readServerCert()
	if err != nil {
		panic(err)
	}
	srv := server.NewServer(version, stream,
		server.ListenAddress(listen),
		//server.Credentials(credentials.NewTLS(tlsConfig)),
		server.MutualTLS(cert, config.ReadCACerts(viper.GetStringSlice("client_ca")...)),
	)
	log.L().Info().Str("listen-addr", listen).Int("pid", os.Getpid()).Msg("Server starting")
	if err := srv.Start(); err != nil {
		panic(err)
	}
}

func executeListBackends(cmd *cobra.Command, args []string) {
	fmt.Println("EventStore Backends:")
	for _, backend := range backends.EventStores() {
		fmt.Printf(" - %s\n", strings.Join(backend.Names(), ", "))
	}
	fmt.Println()
	fmt.Println("SequenceStore Backends:")
	for _, backend := range backends.SequenceStores() {
		fmt.Printf(" - %s\n", strings.Join(backend.Names(), ", "))
	}

}

func readServerCert() (tls.Certificate, error) {
	return tls.LoadX509KeyPair(viper.GetString("tls_cert"), viper.GetString("tls_key"))
}
