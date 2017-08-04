package sourcelogger

import (
	"context"
	"encoding/json"
	"io"
	"os"
	"time"

	"io/ioutil"

	"github.com/apex/log"
	"github.com/segmentio/source-runner/domain"
)

// Operation is an operation string of the action performed
type Operation string

// operations
const (
	SourceStarted      = "source started"
	SourceFinished     = "source finished"
	CollectionStarted  = "collection started"
	CollectionFinished = "collection finished"
	RequestSent        = "request sent"
	ResponseReceived   = "response received"
	ErrorEncountered   = "error encountered"
	SetDispatched      = "set call dispatched"
)

// Metadata defines the log type
type Metadata map[string]interface{}

// Logger is an instance which handles all source logging operations
// and abstracts away the complex bits
// NOTE: more metadata is added server side by source-runner for simplicity.
type Logger struct {
	source  string
	version string
	client  domain.SourceClient
}

// New returns a new Logger instance for use.
func New(source, version string, client domain.SourceClient) *Logger {
	return &Logger{
		source:  source,
		version: version,
		client:  client,
	}
}

// CollectionStarted logs the start of a source collection starting to sync
func (l *Logger) CollectionStarted(collection string) {
	l.write(collection, "info", CollectionStarted, "", nil, nil)
}

// CollectionFinished logs the completions of a source collection
func (l *Logger) CollectionFinished(collection string) {
	l.write(collection, "info", CollectionFinished, "", nil, nil)
}

// RequestSent logs the request address and metadata(optional)
// some considerations for metadata cursor start, end, next or
// query - represents the URL, TCP or database query used to fetch the data
// metadata - any additionaly query options/values that aren't already attached to the query, such as
//            http headers or any extra data that may be pertinent for debugging later specific to the source
func (l *Logger) RequestSent(collection string, query string, metadata Metadata) {
	if metadata == nil {
		metadata = make(Metadata, 1)
	}
	metadata["query"] = query
	l.write(collection, "info", RequestSent, "", metadata, nil)
}

// ResponseReceived logs the request's response information, including raw payload
// query - represents the URL, TCP or database query used to fetch the data
// metadata - any additionaly query options/values that aren't already attached to the query, such as
//            http headers or any extra data that may be pertinent for debugging later specific to the source
func (l *Logger) ResponseReceived(collection string, query string, metadata Metadata, latency time.Duration, payload interface{}) {
	if metadata == nil {
		metadata = make(Metadata, 2)
	}
	metadata["query"] = query
	metadata["request_latency"] = latency.Nanoseconds() / 1e6 // milliseconds
	l.write(collection, "info", ResponseReceived, "", metadata, payload)
}

// Error logs an error that has occured in the source
// NOTE: if an error occurs during a set operation, the error is automatically logged already.
func (l *Logger) Error(collection string, operation string, err error) {
	metadata := Metadata{"operation": operation, "message": err.Error()}
	if errFielder, ok := err.(log.Fielder); ok {
		metadata["fields"] = errFielder.Fields()
	}
	l.write(collection, "error", ErrorEncountered, "", metadata, nil)
}

func (l *Logger) write(collection, level string, operation Operation, id string, attributes Metadata, payload interface{}) {
	var err error
	var b []byte
	var filename, attrs string

	if payload != nil {
		// NOTE: some sources may write results directly to disk before processing, eg. salesforce,
		// in that case we will have a check for what type payload is and process accordingly.
		switch payload.(type) {
		case *os.File:
			f := payload.(*os.File)
			filename = writeFile(f)
		case string:
			filename = writePayload(payload.(string))
		default:
			filename = writeObject(payload)
		}
	}

	if attributes != nil {
		b, err = json.Marshal(attributes)
		if err != nil {
			log.WithError(err).WithField("attributes", attributes).Error("failed to marshal source log attributes")
		}
		attrs = string(b)
	}

	if _, err := l.client.LogSourceEntry(context.Background(), &domain.LogRequest{
		Source: l.source,
		// Version:    l.version, // source-logger doesn't know the version, until it does this is uneeded.
		Collection: collection,
		Level:      level,
		Timestamp:  time.Now().Unix(),
		Operation:  string(operation),
		Attributes: attrs,
		Id:         id,
		Filename:   filename,
	}); err != nil {
		log.WithError(err).Warn("client - failed to log source entry")
	}
}

func writeFile(f *os.File) string {
	f.Seek(0, 0)
	defer f.Seek(0, 0) // to ensure caller can start using file again, if needed.

	file, err := ioutil.TempFile("", "")
	if err != nil {
		log.WithError(err).Error("failed to write temp payload file")
		return ""
	}
	defer file.Close()

	if _, err = io.Copy(file, f); err != nil {
		log.WithError(err).Error("failed to write file payload contents to gzipped file")
		return ""
	}
	return file.Name()
}

func writeObject(payload interface{}) string {
	file, err := ioutil.TempFile("", "")
	if err != nil {
		log.WithError(err).Error("failed to write temp payload file")
		return ""
	}
	defer file.Close()

	enc := json.NewEncoder(file)
	if err = enc.Encode(payload); err != nil {
		log.WithError(err).WithField("payload", payload).Error("failed to marshal source log payload")
	}
	return file.Name()
}

func writePayload(payload string) string {
	file, err := ioutil.TempFile("", "")
	if err != nil {
		log.WithError(err).Error("failed to write temp payload file")
		return ""
	}
	defer file.Close()

	file.Write([]byte(payload))
	return file.Name()
}
