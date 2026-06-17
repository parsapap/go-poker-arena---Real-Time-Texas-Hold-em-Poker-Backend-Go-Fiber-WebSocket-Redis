package metrics

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP metrics
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)
	
	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
	
	// WebSocket metrics
	WebSocketConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "websocket_connections_active",
			Help: "Number of active WebSocket connections",
		},
	)
	
	WebSocketMessagesTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "websocket_messages_total",
			Help: "Total number of WebSocket messages",
		},
		[]string{"type"},
	)
	
	// Game metrics
	ActiveGames = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "poker_games_active",
			Help: "Number of active poker games",
		},
	)
	
	GamesTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "poker_games_total",
			Help: "Total number of poker games started",
		},
	)
	
	PlayersInQueue = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "matchmaking_queue_size",
			Help: "Number of players in matchmaking queue",
		},
	)
	
	HandsDealt = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "poker_hands_dealt_total",
			Help: "Total number of poker hands dealt",
		},
	)
	
	// Room metrics
	ActiveRooms = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "poker_rooms_active",
			Help: "Number of active poker rooms",
		},
	)

	// Error metrics
	ErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Total number of errors by component and type",
		},
		[]string{"component", "type"},
	)

	HTTPErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_errors_total",
			Help: "Total number of HTTP responses with status >= 400",
		},
		[]string{"method", "path", "status"},
	)
)

// RecordError increments the structured error counter for a component.
func RecordError(component, errType string) {
	ErrorsTotal.WithLabelValues(component, errType).Inc()
}

func init() {
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestDuration,
		WebSocketConnections,
		WebSocketMessagesTotal,
		ActiveGames,
		GamesTotal,
		PlayersInQueue,
		HandsDealt,
		ActiveRooms,
		ErrorsTotal,
		HTTPErrorsTotal,
	)
}

// MetricsHandler returns Prometheus metrics handler
func MetricsHandler() fiber.Handler {
	return adaptor.HTTPHandler(promhttp.Handler())
}

// MetricsMiddleware tracks HTTP request metrics
func MetricsMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Use the matched route pattern (e.g. /api/rooms/:id) rather than the
		// raw path so high-cardinality IDs don't explode label cardinality.
		route := c.Route().Path
		if route == "" {
			route = c.Path()
		}

		timer := prometheus.NewTimer(HTTPRequestDuration.WithLabelValues(
			c.Method(),
			route,
		))
		defer timer.ObserveDuration()

		err := c.Next()

		status := c.Response().StatusCode()
		// Convert the numeric HTTP status to its decimal string form (e.g. 200 -> "200").
		// Using strconv.Itoa here is important: string(rune(status)) would instead
		// produce the Unicode character for that code point, corrupting the metric label.
		statusStr := strconv.Itoa(status)
		HTTPRequestsTotal.WithLabelValues(
			c.Method(),
			route,
			statusStr,
		).Inc()

		if status >= 400 {
			HTTPErrorsTotal.WithLabelValues(c.Method(), route, statusStr).Inc()
		}

		return err
	}
}
