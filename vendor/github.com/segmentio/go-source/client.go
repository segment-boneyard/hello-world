package source

import "github.com/segmentio/analytics-go"

// SetMessage is used when using BatchSet
type SetMessage struct {
	Collection string
	ID         string
	Properties map[string]interface{}
}

type GetContextOptions struct {
	AllowFailed bool
}

// Client wraps calls to the RPC service exposed by the source-runner in a Go API.
// Use `New` to create a client.
type Client interface {
	// Set an object with the given collection, id and properties.
	Set(collection string, id string, properties map[string]interface{}) error

	// Set multiple objects, each with a given collection, id and properties.
	SetBatch([]*SetMessage) error

	// Track an object.
	Track(track *analytics.Track) error

	// Identify an object.
	Identify(identify *analytics.Identify) error

	// Group an object.
	Group(group *analytics.Group) error

	GetContext(GetContextOptions) ([]byte, error)
	GetContextIntoFile(GetContextOptions) (string, error)

	SetContext([]byte) error
	SetContextFromFile(string) error

	// Report an error, with an optional collection.
	ReportError(message, collection string) error

	// Report a warning, with an optional collection.
	ReportWarning(message, collection string) error

	// Proxies Statsd Increment
	StatsIncrement(name string, value int64, tags []string) error

	// Proxies Statsd Histogram
	StatsHistogram(name string, value int64, tags []string) error

	// Proxies Statsd Gauge
	StatsGauge(name string, value int64, tags []string) error

	// Proxies KeepAlive
	KeepAlive() error
}

// Config wraps options that can be passed to the client.
type Config struct {
	URL string // URL of the source-runner RPC service.
}

// New creates a client instance with the given configuration.
func New(config *Config) (Client, error) {
	return newClient(config)
}
