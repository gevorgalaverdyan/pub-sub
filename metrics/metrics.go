package metrics

import "github.com/prometheus/client_golang/prometheus"

type Metrics struct {
	UserEvents     prometheus.Counter
	CommerceEvents prometheus.Counter
	ErrorsC        prometheus.CounterVec
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		UserEvents: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "user_events_total",
				Help: "Total number of valid user events processed",
			},
		),
		CommerceEvents: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "commerce_events_total",
				Help: "Total number of valid commerce events processed",
			},
		),
		ErrorsC: *prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "event_errors_total",
				Help: "Total number of event processing errors",
			},
			[]string{"stream", "stage"},
		),
	}
	reg.MustRegister(m.UserEvents)
	reg.MustRegister(m.CommerceEvents)
	reg.MustRegister(m.ErrorsC)
	return m
}
