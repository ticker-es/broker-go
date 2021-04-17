module github.com/ticker-es/broker-go

go 1.16

replace github.com/ticker-es/client-go => ../client-go

require (
	github.com/golang/protobuf v1.5.2
	github.com/mtrense/soil v0.4.0
	github.com/onsi/ginkgo v1.16.0
	github.com/onsi/gomega v1.11.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.7.1
	github.com/ticker-es/client-go v0.0.0
	google.golang.org/grpc v1.37.0
	google.golang.org/protobuf v1.26.0
)
