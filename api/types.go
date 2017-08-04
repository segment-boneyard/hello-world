package api

import (
	"context"
	"github.com/segmentio/go-source/source-logger"
	"net/http"
	"net/url"
	"time"
)

type Object map[string]interface{}

type Request struct {
	Url           string
	Qs            url.Values
	Headers       http.Header
	LogCollection string
}

type ObjectList struct {
	Objects []Object
	HasMore bool
}

type Client interface {
	GetList(context.Context, *Request) (*ObjectList, error)
	GetObject(context.Context, *Request) (Object, error)
}

type HttpClient interface {
	Do(*http.Request) (*http.Response, error)
}

type listResponse struct {
	Object  string   `json:"object"`
	Data    []Object `json:"data"`
	HasMore bool     `json:"has_more"`
}

type ClientOptions struct {
	Secret       string
	BaseUrl      string
	HttpClient   HttpClient
	MaxRps       int
	SourceLogger SourceLogger
}

type SourceLogger interface {
	RequestSent(collection string, query string, metadata sourcelogger.Metadata)
	ResponseReceived(collection string, query string, metadata sourcelogger.Metadata, latency time.Duration, payload interface{})
}
