package controllers

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/lissteron/demoserver/config"
	"github.com/lissteron/demoserver/internal/pkg/client"
)

type RequestHandler struct {
	logger       *log.Logger
	maxRequests  chan struct{}
	urlsLimit    int
	maxInputSize int64
	client       *http.Client
}

func NewRequestHandler(
	logger *log.Logger,
	maxRequests int,
	maxInputSize int64,
	urlsLimit int,
) *RequestHandler {
	return &RequestHandler{
		logger:       logger,
		maxRequests:  make(chan struct{}, maxRequests),
		client:       new(http.Client),
		maxInputSize: maxInputSize,
		urlsLimit:    urlsLimit,
	}
}

func (rh *RequestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	select {
	case rh.maxRequests <- struct{}{}:
	default:
		w.WriteHeader(http.StatusTooManyRequests)

		return
	}

	defer func() {
		<-rh.maxRequests
	}()

	var (
		urls    []string
		decoder = json.NewDecoder(http.MaxBytesReader(w, r.Body, rh.maxInputSize))
	)

	if err := decoder.Decode(&urls); err != nil {
		rh.logger.Printf("[error] unmarshal: %v\n", err)
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	if len(urls) > rh.urlsLimit {
		rh.logger.Printf("[error] to many urls: %d\n", len(urls))
		w.WriteHeader(http.StatusBadRequest)

		return
	}

	cl := client.NewClient(
		rh.logger,
		rh.client,
		config.RequestLimit,
		config.RequestTimeout,
		config.MaxBodySize,
	)

	resp, err := cl.Get(r.Context(), urls)
	if err != nil {
		rh.logger.Printf("[error] get data: %v\n", err)
		w.WriteHeader(http.StatusInternalServerError)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(w)

	if err := encoder.Encode(resp); err != nil {
		rh.logger.Printf("[error] write response: %v\n", err)
	}
}
