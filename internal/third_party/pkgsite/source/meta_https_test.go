// Copyright 2026 Farcloser
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package source //nolint:testpackage // fetchMeta is unexported

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
)

// TestFetchMetaHTTPSOnly: a failed https lookup is an error, never a second
// lookup over plain http, whose answer could be anyone's.
func TestFetchMetaHTTPSOnly(t *testing.T) {
	t.Parallel()

	var calls atomic.Int32

	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		http.Error(w, "unavailable", http.StatusServiceUnavailable)
	}))

	// A plain-http request to this TLS server shows up here as a handshake error.
	var serverLog bytes.Buffer

	srv.Config.ErrorLog = slog.NewLogLogger(slog.NewTextHandler(&serverLog, nil), slog.LevelError)
	srv.StartTLS()
	t.Cleanup(srv.Close)

	client := &Client{httpClient: srv.Client()}
	importPath := strings.TrimPrefix(srv.URL, "https://") + "/mod"

	if meta, err := fetchMeta(t.Context(), client, importPath); err == nil {
		t.Fatalf("fetchMeta(%q) = %+v, want an error", importPath, meta)
	}

	if got := calls.Load(); got != 1 {
		t.Errorf("https lookups = %d, want 1", got)
	}

	if serverLog.Len() != 0 {
		t.Errorf("a plain-http lookup reached the server: %s", serverLog.String())
	}
}
