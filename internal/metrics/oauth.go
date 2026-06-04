package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// OAuthMetrics tracks the OAuth flows, labelled by provider where applicable.
// For exchanges the "outcome" is signed_in / registration_required on success,
// otherwise the service error code; for the others it is the error code or
// success.
type OAuthMetrics struct {
	ExchangesTotal   *prometheus.CounterVec
	CompletionsTotal *prometheus.CounterVec
	LinksTotal       *prometheus.CounterVec
	UnlinksTotal     *prometheus.CounterVec
}

func newOAuthMetrics(reg prometheus.Registerer) *OAuthMetrics {
	factory := promauto.With(reg)

	return &OAuthMetrics{
		ExchangesTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "oauth",
				Name:      "exchanges_total",
				Help:      "Total number of OAuth provider exchanges.",
			},
			[]string{"provider", "outcome"},
		),
		CompletionsTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "oauth",
				Name:      "completions_total",
				Help:      "Total number of OAuth-initiated registration completions.",
			},
			[]string{"outcome"},
		),
		LinksTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "oauth",
				Name:      "links_total",
				Help:      "Total number of OAuth identity links to existing accounts.",
			},
			[]string{"provider", "outcome"},
		),
		UnlinksTotal: factory.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: Namespace,
				Subsystem: "oauth",
				Name:      "unlinks_total",
				Help:      "Total number of OAuth identity unlinks.",
			},
			[]string{"provider", "outcome"},
		),
	}
}

func (m *OAuthMetrics) IncExchange(provider, outcome string) {
	if m == nil {
		return
	}
	m.ExchangesTotal.WithLabelValues(provider, outcome).Inc()
}

func (m *OAuthMetrics) IncCompletion(outcome string) {
	if m == nil {
		return
	}
	m.CompletionsTotal.WithLabelValues(outcome).Inc()
}

func (m *OAuthMetrics) IncLink(provider, outcome string) {
	if m == nil {
		return
	}
	m.LinksTotal.WithLabelValues(provider, outcome).Inc()
}

func (m *OAuthMetrics) IncUnlink(provider, outcome string) {
	if m == nil {
		return
	}
	m.UnlinksTotal.WithLabelValues(provider, outcome).Inc()
}
