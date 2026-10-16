package gitea

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type retryFixture func(*http.Request) (*http.Response, error)

func (f retryFixture) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestReadRetryIsBoundedAndPreservesWriteFailures(t *testing.T) {
	for _, method := range []string{"GET", "HEAD", "POST", "PATCH", "DELETE", "PUT"} {
		for _, code := range []int{200, 400, 401, 403, 405, 409, 429, 500, 502, 503, 504} {
			t.Run(method+http.StatusText(code), func(t *testing.T) {
				calls := 0
				tr := readRetryTransport{base: retryFixture(func(*http.Request) (*http.Response, error) {
					calls++
					return &http.Response{StatusCode: code, Body: io.NopCloser(strings.NewReader("fixture"))}, nil
				})}
				req, _ := http.NewRequestWithContext(context.Background(), method, "https://example.invalid/fixture", nil)
				response, err := tr.RoundTrip(req)
				if err != nil {
					t.Fatal(err)
				}
				defer response.Body.Close()
				want := 1
				if (method == "GET" || method == "HEAD") && (code == 502 || code == 503 || code == 504) {
					want = 2
				}
				if calls != want || response.StatusCode != code {
					t.Fatalf("calls=%d want%d status=%d", calls, want, response.StatusCode)
				}
			})
		}
	}
}

func TestReadRetryHonorsCancellationAndNeverReplaysNetworkFailure(t *testing.T) {
	for _, mode := range []string{"cancelled", "deadline", "network", "body"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
			defer cancel()
			calls := 0
			networkErr := errors.New("fixture transport failure")
			tr := readRetryTransport{base: retryFixture(func(*http.Request) (*http.Response, error) {
				calls++
				if mode == "cancelled" {
					cancel()
				}
				if mode == "network" {
					return nil, networkErr
				}
				return &http.Response{StatusCode: http.StatusServiceUnavailable, Body: io.NopCloser(strings.NewReader("fixture"))}, nil
			})}
			var body io.Reader
			if mode == "body" {
				body = strings.NewReader("not replayable")
			}
			req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.invalid/fixture", body)
			response, err := tr.RoundTrip(req)
			if response != nil {
				defer response.Body.Close()
			}
			if calls != 1 {
				t.Fatalf("attempts=%d", calls)
			}
			if mode == "deadline" && !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("deadline err=%v", err)
			}
			if mode == "network" && !errors.Is(err, networkErr) {
				t.Fatalf("network err=%v", err)
			}
		})
	}
}
