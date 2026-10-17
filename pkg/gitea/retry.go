package gitea

import (
	"io"
	"net/http"
	"time"
)

// readRetryTransport retries one transient gateway response for GET/HEAD only.
// HTTP 500, authorization/conflict failures, network errors and every write
// are returned unchanged: a failed write response may follow a committed write.
type readRetryTransport struct{ base http.RoundTripper }

func (t readRetryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	response, err := t.base.RoundTrip(req)
	if err != nil || response == nil || req.Context().Err() != nil ||
		(req.Method != http.MethodGet && req.Method != http.MethodHead) ||
		(response.StatusCode != http.StatusBadGateway && response.StatusCode != http.StatusServiceUnavailable && response.StatusCode != http.StatusGatewayTimeout) {
		return response, err
	}
	// Do not replay unusual GETs carrying a body, or follow an arbitrary delay
	// suggested by the server. The single retry stays under the caller's deadline.
	if req.Body != nil && req.Body != http.NoBody {
		return response, nil
	}
	_, _ = io.CopyN(io.Discard, response.Body, 64*1024)
	_ = response.Body.Close()
	timer := time.NewTimer(250 * time.Millisecond)
	defer timer.Stop()
	select {
	case <-req.Context().Done():
		return nil, req.Context().Err()
	case <-timer.C:
	}
	return t.base.RoundTrip(req.Clone(req.Context()))
}
