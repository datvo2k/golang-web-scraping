package cache

import (
	"bufio"
	"bytes"
	"github.com/datvo2k/web-scraping/cache/memorycache"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"strings"
	"testing"
	"time"
)

type Policy int

// Transport is an implementation of http.RoundTripper that will return values from a cache
// where possible (avoiding a network request) and will additionally add validators (etag/if-modified-since)
// to repeated requests allowing servers to return 304 / Not Modified
type Transport struct {
	Policy Policy
	// The RoundTripper interface actually used to make requests
	// If nil, http.DefaultTransport is used
	Transport http.RoundTripper
	Cache     Cache
	// If true, responses returned from the cache will be given an extra header, X-From-Cache
	MarkCachedResponses bool
}

const (
	// This policy has no awareness of any HTTP Cache-Control directives.
	// Every request and its corresponding response are cached.
	// When the same request is seen again, the response is returned without transferring anything from the Internet.

	// The Dummy policy is useful for testing spiders faster (without having to wait for downloads every time)
	// and for trying your spider offline, when an Internet connection is not available.
	// The goal is to be able to “replay” a spider run exactly as it ran before.
	Dummy Policy = iota

	// This policy provides a RFC2616 compliant HTTP cache, i.e. with HTTP Cache-Control awareness,
	// aimed at production and used in continuous runs to avoid downloading unmodified data
	// (to save bandwidth and speed up crawls).
	RFC2616
)

const (
	stale = iota
	fresh
	transparent
	XFromCache = "X-From-Cache"
)

type Cache interface {
	Get(key string) (responseBytes []byte, ok bool)
	Set(key string, responseBytes []byte)
	Delete(key string)
}

// cacheKey returns the cache key for req.
func cacheKey(req *http.Request) string {
	if req.Method == http.MethodGet {
		return req.URL.String()
	} else {
		return req.Method + " " + req.URL.String()
	}
}

// CachedResponse returns the cached http.Response for req if present, and nil
// otherwise
func CachedResponse(c Cache, req *http.Request) (resp *http.Response, err error) {
	cachedVal, ok := c.Get(cacheKey(req))
	if !ok {
		return
	}
	b := bytes.NewBuffer(cachedVal)
	return http.ReadResponse(bufio.NewReader(b), req)
}

func NewTransport(c Cache) *Transport {
	return &Transport{
		Policy:              RFC2616,
		Cache:               c,
		MarkCachedResponses: true,
	}
}
