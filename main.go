package main

import (
	"context"
	"github.com/apex/log"
	"github.com/apex/log/handlers/json"
	"github.com/pkg/errors"
	"github.com/segmentio/stats/datadog"
	"github.com/segmentio/conf"
	"github.com/segmentio/go-source"
	"net/http"
	"os"
	"strings"
	"time"
	"github.com/segmentio/ecs-logs-go/log"
	"github.com/segmentio/ecs-logs-go/apex"
	"github.com/segmentio/stats"
	stdlog "log"
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

func setupStats(cfg *config) {Program, stats.Discard, []stats.Tag{
	{Name: "program", Value: Program},
	stats.DefaultEngine = stats.NewEngine(
		{Name: "version", Value: Version},
	}...)
	stats.Register(datadog.NewClient(cfg.DatadogAddr))
}

type config struct {
	Secret          string
	SetTransferId   bool
	DisableAccounts bool
	Rps             int
	DatadogAddr     string
	LogLevel        string
}

func parseConfig() *config {
	rawCfg := struct {
		Secret          string `conf:"secret"`
		SetTransferId   string `conf:"set-transfer-id"`
		DisableAccounts string `conf:"disable-accounts"`
		Rps             int    `conf:"rps"`
	}{Rps: 80}

	conf.LoadWith(&rawCfg, conf.Loader{
		Name:    Program,
		Args:    os.Args[1:],
		Sources: []conf.Source{conf.NewEnvSource("", os.Environ()...)},
	})

	setTransferId := strings.ToLower(rawCfg.SetTransferId)
	disableAccounts := strings.ToLower(rawCfg.DisableAccounts)
	return &config{
		Secret:          rawCfg.Secret,
		Rps:             rawCfg.Rps,
		SetTransferId:   setTransferId == "1" || setTransferId == "yes" || setTransferId == "true",
		DisableAccounts: disableAccounts == "1" || disableAccounts == "yes" || disableAccounts == "true",
		DatadogAddr: "127.0.0.1:8125",
		LogLevel: "INFO",
	}
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

	// initialize api client
	apiClient := api.NewClient(&api.ClientOptions{
		Secret:       cfg.Secret,
		BaseUrl:      "helloworld", //TODO: We don't need an API client yet, but we will
		HttpClient:   &http.Client{Timeout: time.Minute * 5},
		MaxRps:       cfg.Rps,
		SourceClient: sourceClient,
	})

	if cfg.Secret == "" {
		errorMsg := "Invalid credentials (no credentials found)"
		log.Error(errorMsg)
		sourceClient.Log().Error("", "authentication", errors.New(errorMsg))
		sourceClient.ReportError(errorMsg, "")
		return
	}

	// TODO: API test

	// run dispatcher
	d := initDispatcher(apiClient, sourceClient, cfg)
	if err := d.Run(); err != nil {
		log.WithError(err).Fatal("Run failed")
	}

	d.Close()

}

func initDispatcher(apiClient api.Client, sourceClient source.Client, cfg *config) *integration.Dispatcher {
	d := integration.NewDispatcher(sourceClient)

	if !cfg.DisableAccounts {
		d.Register(resource.NewAccount(apiClient))
	}

	d.Register(bundle.New(apiClient,
		resource.NewTransfer(apiClient, cfg.SetTransferId),
		resource.NewTransferReversal(apiClient),
	))

	d.Register(bundle.New(apiClient,
		resource.NewCharge(apiClient),
		resource.NewRefund(apiClient),
		resource.NewCard(apiClient),
		resource.NewBankAccount(apiClient),
	))

	d.Register(bundle.New(apiClient,
		resource.NewSubscription(apiClient),
		resource.NewSubscriptionItem(apiClient),
		resource.NewPlan(apiClient),
		resource.NewInvoice(apiClient),
		resource.NewInvoiceLine(apiClient),
		resource.NewDiscount(apiClient),
		resource.NewCoupon(apiClient),
	))

	d.Register(bundle.New(apiClient,
		resource.NewOrder(apiClient),
		resource.NewOrderShippingMethod(apiClient),
	))

	d.Register(bundle.New(apiClient,
		resource.NewApplicationFee(apiClient),
		resource.NewApplicationFeeRefund(apiClient),
	))

	d.Register(resource.NewBalanceTransaction(apiClient, cfg.SetTransferId))
	d.Register(resource.NewBalanceTransactionFeeDetail(apiClient))
	d.Register(resource.NewCustomer(apiClient))
	d.Register(resource.NewInvoiceItem(apiClient))
	d.Register(resource.NewDispute(apiClient))
	d.Register(resource.NewProduct(apiClient))
	d.Register(resource.NewSku(apiClient))
	d.Register(resource.NewOrderReturn(apiClient))

	return d
}
