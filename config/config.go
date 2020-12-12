package config

import "time"

const (
	HTTPPort              = 8080
	ServerShutdownTimeout = time.Minute * 3

	URLsLimit        = 20
	MaxInputRequests = 100

	MaxBodySize    = 5 << (10 * 2)
	RequestTimeout = time.Second
	RequestLimit   = 4
)
