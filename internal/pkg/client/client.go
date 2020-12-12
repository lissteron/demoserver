package client

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"
)

var ErrBadResponseCode = errors.New("bad response code")

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type Response struct {
	URL  string `json:"url"`
	Body []byte `json:"body"`
}

type Client struct {
	http           HTTPClient
	logger         *log.Logger
	requestTimeout time.Duration
	requestLimit   chan struct{}
	maxBodySize    int64
	wg             sync.WaitGroup
}

func NewClient(
	logger *log.Logger,
	client HTTPClient,
	requestLimit int,
	requestTimeout time.Duration,
	maxBodySize int64,
) *Client {
	return &Client{
		logger:         logger,
		http:           client,
		requestLimit:   make(chan struct{}, requestLimit),
		requestTimeout: requestTimeout,
		maxBodySize:    maxBodySize,
	}
}

func (client *Client) Get(ctx context.Context, urls []string) ([]*Response, error) {
	if err := checkURLs(urls); err != nil {
		return nil, err
	}

	var (
		resp         = make([]*Response, len(urls))
		requestError error
	)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for i, gettedURL := range urls {
		if !client.strartRequest(ctx) {
			break
		}

		go func(i int, gettedURL string) {
			defer client.endRequest()

			body, err := client.download(ctx, gettedURL)
			if err != nil {
				client.logger.Printf("[error] download page: %v", err)

				if requestError == nil {
					requestError = err
				}

				cancel()

				return
			}

			resp[i] = &Response{
				URL:  gettedURL,
				Body: body,
			}
		}(i, gettedURL)
	}

	client.wg.Wait()

	return resp, requestError
}

func (client *Client) download(ctx context.Context, gettedURL string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, client.requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", gettedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("make request error: %w", err)
	}

	resp, err := client.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request error: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %d", ErrBadResponseCode, resp.StatusCode)
	}

	body, err := ioutil.ReadAll(io.LimitReader(resp.Body, client.maxBodySize))
	if err != nil {
		return nil, fmt.Errorf("read body error: %w", err)
	}

	return body, nil
}

func (client *Client) strartRequest(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return false
	default:
	}

	client.wg.Add(1)

	client.requestLimit <- struct{}{}

	return true
}

func (client *Client) endRequest() {
	client.wg.Done()

	<-client.requestLimit
}

func checkURLs(urls []string) error {
	for _, u := range urls {
		if _, err := url.Parse(u); err != nil {
			return fmt.Errorf("bad url: '%s': %w", u, err)
		}
	}

	return nil
}
