package api

import (
	"context"
	"net/http"
	"net/url"
)

type Object map[string]interface{}

type Request struct {
	Url     string
	Qs      url.Values
	Headers http.Header
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
	Secret     string
	BaseUrl    string
	HttpClient HttpClient
	MaxRps     int
}
