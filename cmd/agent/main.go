package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"clustercost-agent-k8s/internal/aggregator"
	"clustercost-agent-k8s/internal/collector"
	"clustercost-agent-k8s/internal/config"
	"clustercost-agent-k8s/internal/enricher"
	"clustercost-agent-k8s/internal/exporter"
	"clustercost-agent-k8s/internal/kube"
	"clustercost-agent-k8s/internal/logging"
	"clustercost-agent-k8s/internal/pricing"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	logger := logging.New(cfg.LogLevel)

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer cancel()

	kubeClient, err := kube.NewClient(cfg.ClusterName, cfg.KubeconfigPath)
	if err != nil {
		logger.Error("failed to create kube client", slog.String("error", err.Error()))
		os.Exit(1)
	}

	clusterCollector := collector.NewClusterCollector(kubeClient, logger)
	metricsCollector := collector.NewMetricsCollector(kubeClient, logger)
	calc := pricing.NewCalculator(cfg.Pricing.CPUCoreHourPriceUSD, cfg.Pricing.MemoryGiBHourPriceUSD)
	labelEnricher := enricher.NewLabelEnricher()
	agg := aggregator.NewCostAggregator(calc, labelEnricher, logger)
	promExporter := exporter.NewPrometheusExporter(cfg.ClusterName)
	api := exporter.NewHTTPAPI(agg)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promExporter.Handler())
	api.Register(mux)

	server := exporter.NewServer(cfg.ListenAddr, mux, logger)

	go runScraper(ctx, clusterCollector, metricsCollector, agg, promExporter, cfg.ScrapeInterval(), logger)

	if err := server.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
		logger.Error("server error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func runScraper(ctx context.Context, clusterCollector *collector.ClusterCollector, metricsCollector *collector.MetricsCollector, agg *aggregator.CostAggregator, promExporter *exporter.PrometheusExporter, interval time.Duration, logger *slog.Logger) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		if err := scrapeOnce(ctx, clusterCollector, metricsCollector, agg, promExporter, logger); err != nil {
			logger.Warn("scrape failed", slog.String("error", err.Error()))
		}

		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func scrapeOnce(ctx context.Context, clusterCollector *collector.ClusterCollector, metricsCollector *collector.MetricsCollector, agg *aggregator.CostAggregator, promExporter *exporter.PrometheusExporter, logger *slog.Logger) error {
	scrapeCtx, cancel := context.WithTimeout(ctx, 25*time.Second)
	defer cancel()

	snapshot, err := clusterCollector.Collect(scrapeCtx)
	if err != nil {
		return err
	}

	metrics, err := metricsCollector.CollectPodMetrics(scrapeCtx)
	if err != nil {
		logger.Warn("using requests due to metrics error", slog.String("error", err.Error()))
		metrics = map[string]kube.PodUsage{}
	}

	data := agg.Aggregate(snapshot, metrics)
	promExporter.Update(data)

	return nil
}
