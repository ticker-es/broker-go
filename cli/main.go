package main

import (
	"os"

	. "github.com/mtrense/soil/config"
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
	srv := server.NewServer(listen, version, stream)
	log.L().Info().Str("listenAddr", listen).Int("pid", os.Getpid()).Msg("Server starting")
	if err := srv.Start(); err != nil {
		panic(err)
	}
}
