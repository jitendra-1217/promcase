package promcase

import (
	"fmt"
	"github.com/fatih/structs"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"regexp"
	"strconv"
	"strings"
)

// MessageRawRegex represents a raw Udp message format.
// A raw message is expected to be in following format - "type|name|help|labels|action|args".
// E.g. "h|test_hist|Test histogram||o|v=1.0005,b=1#2#5"
const MessageRawRegex = "^(.*)\\|(.*)\\|(.*)\\|(.*)\\|(.*)\\|(.*)$"

// Message represents instance of a UDP message received.
type Message struct {
	Raw          string
	FromAddress  string
	MetricType   string
	MetricName   string
	MetricHelp   string
	MetricLabels map[string]string
	ActionType   string // E.g. increment, decrement, set, observe etc.
	ActionArgs   map[string]string
}

// NewMessage function creates Message from given raw string and other args.
func NewMessage(raw string, fromAddress string) (Message, error) {
	m := Message{
		Raw:          strings.Trim(raw, "\n"),
		FromAddress:  fromAddress,
		MetricLabels: make(map[string]string),
		ActionArgs:   make(map[string]string)}
	re, err := regexp.Compile(MessageRawRegex)
	if err != nil {
		return m, err
	}
	matches := re.FindStringSubmatch(m.Raw)
	if matches == nil {
		return m, fmt.Errorf("raw message is not in expected format: %s", m.Raw)
	}
	m.MetricType = matches[1]
	m.MetricName = matches[2]
	m.MetricHelp = matches[3]
	for _, s := range strings.Split(matches[4], ",") {
		v := strings.Split(s, "=")
		// Possible value is [""], i.e. no labels provided.
		if len(v) == 2 {
			m.MetricLabels[v[0]] = v[1]
		}
	}
	m.ActionType = matches[5]
	for _, s := range strings.Split(matches[6], ",") {
		v := strings.Split(s, "=")
		// Possible value is [""], i.e. no args provided, in which case it should error.
		if len(v) < 2 {
			return m, fmt.Errorf("raw message is not in expected format, incorrect action args: %s", m.Raw)
		}
		m.ActionArgs[v[0]] = v[1]
	}
	return m, nil
}

// Process method processes the message
func (m Message) Process() error {
	// Deferred recovery from any panics from prometheus client lib.
	defer func() {
		if r := recover(); r != nil {
			log.WithFields(log.Fields{"m": structs.Map(m)}).Error(r)
		}
	}()
	switch m.MetricType {
	case MetricTypeCounter:
		return m.ProcessCounter()
	case MetricTypeGauge:
		return m.ProcessGauge()
	case MetricTypeHistogram:
		return m.ProcessHistogram()
	default:
		return fmt.Errorf("metric type is invalid: %s", m.MetricType)
	}
	return nil
}

// ProcessCounter method handles counter type metric
func (m Message) ProcessCounter() error {
	metric, ok := Register.TypeCounter[m.MetricName]
	if !ok {
		metric = prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: m.MetricName,
				Help: m.MetricHelp,
			},
			m.GetMetricLabelsKeys())
		prometheus.MustRegister(metric)
		Register.TypeCounter[m.MetricName] = metric
	}
	switch m.ActionType {
	case ActionTypeInc:
		arg, ok := m.ActionArgs["v"]
		if !ok {
			arg = "1"
		}
		v, err := strconv.ParseFloat(arg, 64)
		if err != nil {
			return err
		}
		metric.With(m.MetricLabels).Add(v)
	default:
		return fmt.Errorf("action type is invalid for counter: %s", m.ActionType)
	}
	return nil
}

// ProcessGauge method handles gauge type metric
func (m Message) ProcessGauge() error {
	metric, ok := Register.TypeGauge[m.MetricName]
	if !ok {
		metric = prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: m.MetricName,
				Help: m.MetricHelp,
			},
			m.GetMetricLabelsKeys())
		prometheus.MustRegister(metric)
		Register.TypeGauge[m.MetricName] = metric
	}
	arg, ok := m.ActionArgs["v"]
	if !ok {
		return fmt.Errorf("value is not sent for gauge: %s", m.MetricName)
	}
	v, err := strconv.ParseFloat(arg, 64)
	if err != nil {
		return err
	}
	switch m.ActionType {
	case ActionTypeInc:
		metric.With(m.MetricLabels).Add(v)
	case ActionTypeDec:
		metric.With(m.MetricLabels).Sub(v)
	case ActionTypeSet:
		metric.With(m.MetricLabels).Set(v)
	default:
		return fmt.Errorf("action type is invalid for gauge: %s", m.ActionType)
	}
	return nil
}

// ProcessHistogram method handles histogram type metric
func (m Message) ProcessHistogram() error {
	metric, ok := Register.TypeHistogram[m.MetricName]
	if !ok {
		var buckets []float64
		bucketsArg, ok := m.ActionArgs["b"]
		if !ok {
			// If buckets value is not provided in args list then uses prometheus defaults.
			buckets = prometheus.DefBuckets
		} else {
			// Else splits comma separated bucket values and uses that.
			bucketsArgSplits := strings.Split(bucketsArg, "#")
			buckets = make([]float64, len(bucketsArgSplits))
			for i, v := range bucketsArgSplits {
				floatV, err := strconv.ParseFloat(v, 64)
				if err != nil {
					return err
				}
				buckets[i] = floatV
			}
		}
		metric = prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    m.MetricName,
				Help:    m.MetricHelp,
				Buckets: buckets,
			},
			m.GetMetricLabelsKeys())
		prometheus.MustRegister(metric)
		Register.TypeHistogram[m.MetricName] = metric
	}
	arg, ok := m.ActionArgs["v"]
	if !ok {
		return fmt.Errorf("value is not sent for histogram: %s", m.MetricName)
	}
	v, err := strconv.ParseFloat(arg, 64)
	if err != nil {
		return err
	}
	switch m.ActionType {
	case ActionTypeObserve:
		metric.With(m.MetricLabels).Observe(v)
	default:
		return fmt.Errorf("action type is invalid for histogram: %s", m.ActionType)
	}
	return nil
}

func (m Message) GetMetricLabelsKeys() []string {
	keys := make([]string, len(m.MetricLabels))
	i := 0
	for k := range m.MetricLabels {
		keys[i] = k
		i++
	}
	return keys
}
