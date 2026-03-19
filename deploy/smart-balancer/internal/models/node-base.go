package models

import "net/url"

type Protocol string

const (
	HTTP Protocol = "http"
	GRPC Protocol = "grpc"
)

type Node struct {
	URL      url.URL
	Protocol Protocol
	Metrics  interface{}
}
