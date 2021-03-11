package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	nearapi "github.com/bisontrails/near-exporter/client"
	"github.com/bisontrails/near-exporter/collector"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
)

func main() {
	configureEnvironment()
	internalURL := viper.GetString("INTERNAL_URL")
	externalURL := viper.GetString("EXTERNAL_URL")
	accountID := viper.GetString("ACCOUNT_ID")
	listenAddress := viper.GetString("LISTEN_ADDRESS")

	flag.Parse()
	if len(flag.Args()) > 0 {
		flag.Usage()
	}

	client := nearapi.NewClient(internalURL)

	devClient := nearapi.NewClient(externalURL)

	rpcMetricCollector := collector.NewNodeRpcMetrics(client, devClient, accountID)
	fmt.Println("do the thing")
	go rpcMetricCollector.RecordValidators()

	registry := prometheus.NewPedanticRegistry()
	registry.MustRegister(
		rpcMetricCollector,
		collector.NewDevNodeRpcMetrics(devClient),
	)

	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{
		ErrorLog:      log.New(os.Stderr, log.Prefix(), log.Flags()),
		ErrorHandling: promhttp.ContinueOnError,
	})

	http.Handle("/metrics", handler)
	log.Fatal(http.ListenAndServe(listenAddress, nil))
}

func configureEnvironment() {
	viper.AutomaticEnv()
	viper.SetDefault("INTERNAL_URL", "http://localhost:3030")
	viper.SetDefault("EXTERNAL_URL", "https://rpc.betanet.near.org")
	viper.SetDefault("ACCOUNT_ID", "test")
	viper.SetDefault("LISTEN_ADDRESS", ":9333")
	viper.SetDefault("CLIENT_TIMEOUT_SECONDS", 30)
}
