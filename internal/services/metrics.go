package services

import (
	"context"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/push"
	"go.uber.org/fx"
	"os"
	"p0-sink/internal/infrastructure"
	fx_utils "p0-sink/internal/utils/fx"
	"time"
)

type IMetricsService interface {
	IncTotalBlocksSent()
	IncTotalBytesSent(bytes int)
	IncTotalBatchesSent()
	IncErrorsCount()
	IncRetriesCount()
	SetLastBlockSentAt()
	MeasureDownloadLatency(fn func())
	MeasureFilterLatency(fn func())
	MeasureDecompressLatency(fn func())
	MeasureCompressLatency(fn func())
	MeasureSerializeLatency(fn func())
	MeasureSendLatency(fn func())
	MeasureCommitLatency(fn func())
}

type metricsServiceParams struct {
	fx.In

	Logger infrastructure.ILogger
	Config infrastructure.IConfig
}

type metricsService struct {
	ctx      context.Context
	cancel   context.CancelFunc
	hostName string
	logger   infrastructure.ILogger
	config   infrastructure.IConfig
	push     *push.Pusher

	totalBlocksSent  prometheus.Counter
	totalBytesSent   prometheus.Counter
	totalBatchesSent prometheus.Counter
	errorsCount      prometheus.Counter
	retriesCount     prometheus.Counter
	lastBlockSentAt  prometheus.Gauge

	downloadLatency   prometheus.Histogram
	filterLatency     prometheus.Histogram
	decompressLatency prometheus.Histogram
	compressLatency   prometheus.Histogram
	serializeLatency  prometheus.Histogram
	sendLatency       prometheus.Histogram
	commitLatency     prometheus.Histogram
}

func FxMetricsService() fx.Option {
	return fx_utils.AsProvider(newMetricsService, new(IMetricsService))
}

func newMetricsService(lc fx.Lifecycle, params metricsServiceParams) IMetricsService {
	hostname, err := os.Hostname()

	if err != nil {
		panic(fmt.Sprintf("failed to get hostname: %v", err))
	}

	msBucket := []float64{10, 20, 35, 50, 75, 100, 150, 200, 300, 400, 500, 1000}
	ms := &metricsService{
		hostName: hostname,
		logger:   params.Logger,
		config:   params.Config,
		totalBlocksSent: promauto.NewCounter(prometheus.CounterOpts{
			Name: "project_zero:sink:blocks_sent",
			Help: "The total number of blocks sent",
		}),
		totalBytesSent: promauto.NewCounter(prometheus.CounterOpts{
			Name: "project_zero:sink:bytes_sent",
			Help: "The total number of bytes sent",
		}),
		totalBatchesSent: promauto.NewCounter(prometheus.CounterOpts{
			Name: "project_zero:sink:batches_sent",
			Help: "The total number of batches sent",
		}),
		errorsCount: promauto.NewCounter(prometheus.CounterOpts{
			Name: "project_zero:sink:errors_count",
			Help: "The total number of errors",
		}),
		retriesCount: promauto.NewCounter(prometheus.CounterOpts{
			Name: "project_zero:sink:retries_count",
			Help: "The total number of retries",
		}),
		lastBlockSentAt: promauto.NewGauge(prometheus.GaugeOpts{
			Name: "project_zero:sink:last_block_sent_at",
			Help: "The timestamp of the last block sent",
		}),
		downloadLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "project_zero:sink:download_latency",
			Help:    "The latency of the download in milliseconds",
			Buckets: msBucket,
		}),
		filterLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "project_zero:sink:filter_latency",
			Help:    "The latency of the filter in milliseconds",
			Buckets: msBucket,
		}),
		decompressLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "project_zero:sink:decompress_latency",
			Help:    "The latency of the decompress in milliseconds",
			Buckets: msBucket,
		}),
		compressLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "project_zero:sink:compress_latency",
			Help:    "The latency of the compress in milliseconds",
			Buckets: msBucket,
		}),
		serializeLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "project_zero:sink:serialize_latency",
			Help:    "The latency of the serialize in milliseconds",
			Buckets: msBucket,
		}),
		sendLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "project_zero:sink:send_latency",
			Help:    "The latency of the send in milliseconds",
			Buckets: msBucket,
		}),
		commitLatency: promauto.NewHistogram(prometheus.HistogramOpts{
			Name:    "project_zero:sink:commit_latency",
			Help:    "The latency of the commit in milliseconds",
			Buckets: msBucket,
		}),
	}

	lc.Append(fx.Hook{
		OnStart: ms.start,
		OnStop:  ms.stop,
	})

	return ms
}

func (ms *metricsService) IncTotalBlocksSent() {
	ms.totalBlocksSent.Inc()
}

func (ms *metricsService) IncTotalBytesSent(bytes int) {
	ms.totalBytesSent.Add(float64(bytes))
}

func (ms *metricsService) IncTotalBatchesSent() {
	ms.totalBatchesSent.Inc()
}

func (ms *metricsService) IncErrorsCount() {
	ms.errorsCount.Inc()
}

func (ms *metricsService) IncRetriesCount() {
	ms.retriesCount.Inc()
}

func (ms *metricsService) SetLastBlockSentAt() {
	ms.lastBlockSentAt.SetToCurrentTime()
}

func (ms *metricsService) MeasureDownloadLatency(fn func()) {
	ms.measure(fn, ms.downloadLatency)
}

func (ms *metricsService) MeasureFilterLatency(fn func()) {
	ms.measure(fn, ms.filterLatency)
}

func (ms *metricsService) MeasureDecompressLatency(fn func()) {
	ms.measure(fn, ms.decompressLatency)
}

func (ms *metricsService) MeasureCompressLatency(fn func()) {
	ms.measure(fn, ms.compressLatency)
}

func (ms *metricsService) MeasureSerializeLatency(fn func()) {
	ms.measure(fn, ms.serializeLatency)
}

func (ms *metricsService) MeasureSendLatency(fn func()) {
	ms.measure(fn, ms.sendLatency)
}

func (ms *metricsService) MeasureCommitLatency(fn func()) {
	ms.measure(fn, ms.commitLatency)
}

func (ms *metricsService) measure(fn func(), histogram prometheus.Histogram) {
	start := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		fn()
	}))
	duration := start.ObserveDuration()

	histogram.Observe(float64(duration.Milliseconds()))
}

func (ms *metricsService) start(_ context.Context) error {
	ms.lastBlockSentAt.SetToCurrentTime()

	if ms.config.GetMetricsEnabled() {
		ms.push = push.New(ms.config.GetMetricsUrl(), ms.config.GetMetricsJobName()).
			Grouping("streamId", ms.config.GetStreamId()).
			Grouping("instance", ms.hostName).
			Collector(ms.totalBlocksSent).
			Collector(ms.totalBytesSent).
			Collector(ms.totalBatchesSent).
			Collector(ms.errorsCount).
			Collector(ms.retriesCount).
			Collector(ms.lastBlockSentAt).
			Collector(ms.downloadLatency).
			Collector(ms.filterLatency).
			Collector(ms.decompressLatency).
			Collector(ms.compressLatency).
			Collector(ms.serializeLatency).
			Collector(ms.sendLatency).
			Collector(ms.commitLatency)

		ms.ctx, ms.cancel = context.WithCancel(context.Background())
		go ms.startSendingMetricsInterval()
		ms.logger.Info("metrics sends enabled")
		return nil
	} else {
		ms.logger.Info("metrics sends disabled")
		return nil
	}
}

func (ms *metricsService) stop(_ context.Context) error {
	ms.cancel()
	ms.logger.Info("metrics sender stopped")
	return nil
}

func (ms *metricsService) startSendingMetricsInterval() {
	ticker := time.NewTicker(ms.config.GetMetricsInterval())

	defer ticker.Stop()

	for {
		select {
		case <-ms.ctx.Done():
			return
		case <-ticker.C:
			if err := ms.push.Push(); err != nil {
				ms.logger.Error(fmt.Sprintf("sending metrics error: %s", err))
			}
		}
	}
}
