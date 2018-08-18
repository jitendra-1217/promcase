# Promcase
Binary wrapping Prometheus's go client, listens on Udp

## What
This is intended to be light binary which can be run on each instance/pod. And Php like applications which run in fork-process based model via Apache or Nginx, can Udp metrics to this binary. The binary holds the metrics coming from across processes (e.g. Php-fpm thread, cli workers) and exposes to Prometheus for scrape.

## Why
The unofficial client library for Php does not perform well with large number (say > 50k) metrics. It uses either local Redis or Apcu to hold metrics from across processes. In both cases rendering of metrics for scrape is slow.

## Usage

```sh
# Assuming you have go setup in your machine
go install github.com/jitendra-1217/promcase
TCP_PORT=10002 UDP_PORT=10001 promcase
# Now it will listen on udp port 10001 for metrics and exposes metrics on http://0.0.0.0:10002/metrics
```

```sh
# Anatomy of Udp message
# metric-type|metric-name|metric-help-text|labels("," separated key=value)|action-type|action-args
# Pushing metrics of different types with labels
echo "c|counter_metric|A test counter metric|region=south,planet=earth|i|v=1" | nc -u -w0 127.0.0.1 10001
echo "g|gauge_metric|A test gauge metric|region=south,planet=earth|i|v=1.23" | nc -u -w0 127.0.0.1 10001
echo "h|histogram_metric|A test histogram type metric||o|v=1.50,b=1#2#3" | nc -u -w0 127.0.0.1 10001
```

```
# HELP counter_metric A test counter metric
# TYPE counter_metric counter
counter_metric{planet="earth",region="south"} 1
# HELP gauge_metric A test gauge metric
# TYPE gauge_metric gauge
gauge_metric{planet="earth",region="south"} 1.23
# HELP histogram_metric A test histogram type metric
# TYPE histogram_metric histogram
histogram_metric_bucket{le="1"} 0
histogram_metric_bucket{le="2"} 1
histogram_metric_bucket{le="3"} 1
histogram_metric_bucket{le="+Inf"} 1
histogram_metric_sum 1.5
histogram_metric_count 1
```

## Limitations
- Prometheus's official client library have really good interface and extendibility. It is attainable by writing a thin client lib which constructs and forwards Udp messages to this binary.
- I have not fully understood or used multiple registries. This binary for now just understands default registry.
- It only implements most common use cases around counter, gauge and histogram type metric.
