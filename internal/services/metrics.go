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
	IncTotalBlocksSent(num int)
	IncTotalBytesSent(bytes uint64)
	IncTotalBatchesSent()
	IncErrorsCount()
	IncRetriesCount()
	SetLastBlockSentAt()
	MeasureDownloadLatency(fn func() ([]byte, error)) ([]byte, error)
	MeasureFilterLatency(fn func())
	MeasureDecompressLatency(fn func() ([]byte, error)) ([]byte, error)
	MeasureCompressLatency(fn func() ([]byte, error)) ([]byte, error)
	MeasureSerializeLatency(fn func() ([]byte, int, error)) ([]byte, int, error)
	MeasureSendLatency(fn func() error) error
	MeasureCommitLatency(fn func() error) error
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

func (ms *metricsService) IncTotalBlocksSent(num int) {
	ms.totalBlocksSent.Add(float64(num))
}

func (ms *metricsService) IncTotalBytesSent(bytes uint64) {
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

func (ms *metricsService) MeasureDownloadLatency(fn func() ([]byte, error)) ([]byte, error) {
	timer := prometheus.NewTimer(ms.downloadLatency)
	defer func() {
		duration := timer.ObserveDuration()
		ms.downloadLatency.Observe(float64(duration.Milliseconds()))
	}()

	return fn()
}

func (ms *metricsService) MeasureFilterLatency(fn func()) {
	ms.measure(fn, ms.filterLatency)
}

func (ms *metricsService) MeasureDecompressLatency(fn func() ([]byte, error)) ([]byte, error) {
	return ms.measureWithBytes(fn, ms.decompressLatency)
}

func (ms *metricsService) MeasureCompressLatency(fn func() ([]byte, error)) ([]byte, error) {
	return ms.measureWithBytes(fn, ms.compressLatency)
}

func (ms *metricsService) MeasureSerializeLatency(fn func() ([]byte, int, error)) ([]byte, int, error) {
	timer := prometheus.NewTimer(ms.serializeLatency)
	defer func() {
		duration := timer.ObserveDuration()
		ms.serializeLatency.Observe(float64(duration.Milliseconds()))
	}()

	return fn()
}

func (ms *metricsService) MeasureSendLatency(fn func() error) error {
	return ms.measureWithError(fn, ms.sendLatency)
}

func (ms *metricsService) MeasureCommitLatency(fn func() error) error {
	return ms.measureWithError(fn, ms.commitLatency)
}

func (ms *metricsService) measure(fn func(), histogram prometheus.Histogram) {
	start := prometheus.NewTimer(prometheus.ObserverFunc(func(v float64) {
		fn()
	}))
	duration := start.ObserveDuration()

	histogram.Observe(float64(duration.Milliseconds()))
}

func (ms *metricsService) measureWithBytes(fn func() ([]byte, error), histogram prometheus.Histogram) ([]byte, error) {
	timer := prometheus.NewTimer(histogram)
	defer func() {
		duration := timer.ObserveDuration()
		histogram.Observe(float64(duration.Milliseconds()))
	}()

	return fn()
}

func (ms *metricsService) measureWithError(fn func() error, histogram prometheus.Histogram) error {
	timer := prometheus.NewTimer(histogram)
	defer func() {
		duration := timer.ObserveDuration()
		histogram.Observe(float64(duration.Milliseconds()))
	}()

	return fn()
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

		sendCtx, sendCancelFunc := context.WithCancel(context.Background())

		ms.ctx = sendCtx
		ms.cancel = sendCancelFunc

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
