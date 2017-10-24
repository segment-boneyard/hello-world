package main

import (
	"github.com/apex/log"
	"github.com/apex/log/handlers/json"
	"github.com/segmentio/stats"
	"github.com/segmentio/stats/datadog"
	"github.com/segmentio/go-source"
	"os"
	"github.com/segmentio/ecs-logs-go/log"
	"github.com/segmentio/ecs-logs-go/apex"
	stdlog "log"
	"gopkg.in/alecthomas/kingpin.v2"
)


const (
	Program = "hello-world-source"
	Version = "0.0.1"
)

func initLogger() {
	globalLogger := &log.Logger{
		Level:   log.InfoLevel,
		Handler: json.Default,
	}
	log.Log = globalLogger.WithFields(log.Fields{
		"program": Program,
		"version": Version,
	})
}


func setupLogging(cfg *config) {
	handler := log_ecslogs.NewHandler(os.Stdout)
	writer := log_ecslogs.NewWriter("", stdlog.Flags(), handler)
	stdlog.SetOutput(writer)

	log.SetHandler(apex_ecslogs.NewHandler(os.Stdout))
	log.SetLevel(log.MustParseLevel(cfg.LogLevel))
	log.Log = log.WithFields(log.Fields{
		"program": Program,
		"version": Version,
	})
}

func setupStats(cfg *config) {
	stats.DefaultEngine = stats.NewEngine(Program, stats.Discard, []stats.Tag{
		{Name: "program", Value: Program},
		{Name: "version", Value: Version},
	}...)
	stats.Register(datadog.NewClient(cfg.DatadogAddr))
}


type config struct {
	DatadogAddr string
	LogLevel    string
	BindAddr    string
	Message 	string
	CollectionName string
}

func parseConfig() *config {
	cfg := &config{}
	kingpin.Flag("bind-addr", "Address and port to listen on").Default(":3000").StringVar(&cfg.BindAddr)
	kingpin.Flag("datadog-addr", "Datadog statsd host and port").Default("127.0.0.1:8125").StringVar(&cfg.DatadogAddr)
	kingpin.Flag("log-level", "Logging level").Default("INFO").StringVar(&cfg.LogLevel)
	kingpin.Flag("message", "Message to Serve").Default("Hello, World").StringVar(&cfg.Message)
	kingpin.Flag("collection", "Collection Name").Default("helloworld").StringVar(&cfg.CollectionName)
	kingpin.Parse()
	return cfg
}

func main() {
	// Basic Setup
	cfg := parseConfig()
	setupLogging(cfg)
	setupStats(cfg)
	defer stats.Flush()

	// initialize source client
	sourceClient, err := source.New(&source.Config{
		URL: "localhost:4000",
	})
	if err != nil {
		log.WithError(err).Fatal("failed to initialize source client")
	}
	if err := sourceClient.KeepAlive(); err != nil {
		log.WithError(err).Fatal("keepalive call failed")
	}

	var properties map[string]interface{}
	properties = make(map[string]interface{})
	properties["message"] = cfg.Message

	sourceClient.Set(cfg.CollectionName, "1", properties)

}