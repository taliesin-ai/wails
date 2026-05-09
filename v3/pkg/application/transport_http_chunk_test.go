package application

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// newNopLogger returns a slog.Logger that discards all output.
func newNopLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// newTestTransport builds an HTTPTransport ready for unit tests.
// processBodyFn is set to a no-op; callers that need to inspect it should
// replace it after calling this function.
func newTestTransport() *HTTPTransport {
	tr := NewHTTPTransport(HTTPTransportWithLogger(newNopLogger()))
	tr.stopCleanup = make(chan struct{})
	return tr
}

// makeChunkReq builds a POST /wails/runtime request carrying one slice of a
// chunked upload.
func makeChunkReq(chunkID string, index, total int, payload []byte) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/wails/runtime", bytes.NewReader(payload))
	req.Header.Set(chunkIDHeader, chunkID)
	req.Header.Set(chunkIndexHeader, fmt.Sprintf("%d", index))
	req.Header.Set(chunkTotalHeader, fmt.Sprintf("%d", total))
	req.Header.Set("x-wails-client-id", "test-client")
	return req
}

// sendChunks splits payload into nChunks slices and drives them through
// handleChunkedRequest, returning the last httptest.ResponseRecorder.
func sendChunks(tr *HTTPTransport, chunkID string, payload []byte, nChunks int) []*httptest.ResponseRecorder {
	chunkSize := (len(payload) + nChunks - 1) / nChunks
	var rws []*httptest.ResponseRecorder
	for i := 0; i < nChunks; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if end > len(payload) {
			end = len(payload)
		}
		rw := httptest.NewRecorder()
		req := makeChunkReq(chunkID, i, nChunks, payload[start:end])
		tr.handleChunkedRequest(rw, req, chunkID)
		rws = append(rws, rw)
	}
	return rws
}

// TestChunkAccumulation_IntermediateChunksReturn200 checks that all chunks
// except the last return HTTP 200 with an empty body.
func TestChunkAccumulation_IntermediateChunksReturn200(t *testing.T) {
	tr := newTestTransport()
	defer close(tr.stopCleanup)
	tr.processBodyFn = func(_ http.ResponseWriter, _ *http.Request, _ []byte) {}

	payload := bytes.Repeat([]byte("X"), 3*1024*1024) // 3 MB
	const nChunks = 6
	rws := sendChunks(tr, "mid-200", payload, nChunks)

	for i := 0; i < nChunks-1; i++ {
		if rws[i].Code != http.StatusOK {
			t.Errorf("chunk %d: expected 200, got %d", i, rws[i].Code)
		}
		if rws[i].Body.Len() != 0 {
			t.Errorf("chunk %d: expected empty body, got %q", i, rws[i].Body.String())
		}
	}
}

// TestChunkAccumulation_AssemblyCorrect verifies the assembled payload equals
// the original for several payload sizes.
func TestChunkAccumulation_AssemblyCorrect(t *testing.T) {
	cases := []struct {
		name    string
		size    int
		nChunks int
	}{
		{"1KB_1chunk", 1 * 1024, 1},
		{"512KB_2chunks", 512 * 1024, 2},
		{"1MB_2chunks", 1024 * 1024, 2},
		{"2MB_4chunks", 2 * 1024 * 1024, 4},
		{"3MB_6chunks", 3 * 1024 * 1024, 6},
		{"5MB_10chunks", 5 * 1024 * 1024, 10},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tr := newTestTransport()
			defer close(tr.stopCleanup)

			payload := bytes.Repeat([]byte("A"), tc.size)
			var assembled []byte
			tr.processBodyFn = func(_ http.ResponseWriter, _ *http.Request, body []byte) {
				assembled = body
			}

			sendChunks(tr, tc.name, payload, tc.nChunks)

			if !bytes.Equal(assembled, payload) {
				t.Errorf("assembled %d bytes; want %d bytes", len(assembled), len(payload))
			}
		})
	}
}

// TestChunkAccumulation_ChunkStoreCleared checks that the chunkStore entry is
// removed after the final chunk arrives (prevents memory leak).
func TestChunkAccumulation_ChunkStoreCleared(t *testing.T) {
	tr := newTestTransport()
	defer close(tr.stopCleanup)
	tr.processBodyFn = func(_ http.ResponseWriter, _ *http.Request, _ []byte) {}

	payload := bytes.Repeat([]byte("B"), 1024*1024) // 1 MB
	sendChunks(tr, "store-clear", payload, 2)

	if _, ok := tr.chunkStore.Load("store-clear"); ok {
		t.Fatal("chunkStore entry was NOT removed after final chunk arrived")
	}
}

// TestChunkAccumulation_InvalidIndex verifies that a negative or non-numeric
// chunk index produces HTTP 422 immediately.
func TestChunkAccumulation_InvalidIndex(t *testing.T) {
	cases := []struct{ index, total string }{
		{"-1", "3"},
		{"abc", "3"},
	}
	for _, tc := range cases {
		t.Run(tc.index, func(t *testing.T) {
			tr := newTestTransport()
			defer close(tr.stopCleanup)

			req := httptest.NewRequest(http.MethodPost, "/wails/runtime", strings.NewReader("data"))
			req.Header.Set(chunkIDHeader, "bad")
			req.Header.Set(chunkIndexHeader, tc.index)
			req.Header.Set(chunkTotalHeader, tc.total)

			rw := httptest.NewRecorder()
			tr.handleChunkedRequest(rw, req, "bad")

			if rw.Code != http.StatusUnprocessableEntity {
				t.Errorf("expected 422, got %d", rw.Code)
			}
		})
	}
}

// TestChunkAccumulation_InvalidTotal verifies that a zero or non-numeric
// total chunk count produces HTTP 422.
func TestChunkAccumulation_InvalidTotal(t *testing.T) {
	cases := []struct{ index, total string }{
		{"0", "0"},
		{"0", "xyz"},
	}
	for _, tc := range cases {
		t.Run(tc.total, func(t *testing.T) {
			tr := newTestTransport()
			defer close(tr.stopCleanup)

			req := httptest.NewRequest(http.MethodPost, "/wails/runtime", strings.NewReader("data"))
			req.Header.Set(chunkIDHeader, "bad")
			req.Header.Set(chunkIndexHeader, tc.index)
			req.Header.Set(chunkTotalHeader, tc.total)

			rw := httptest.NewRecorder()
			tr.handleChunkedRequest(rw, req, "bad")

			if rw.Code != http.StatusUnprocessableEntity {
				t.Errorf("expected 422, got %d", rw.Code)
			}
		})
	}
}

// TestChunkAccumulation_TTLEviction verifies that sweepExpired removes
// accumulators older than chunkTTL without waiting 30 s.
func TestChunkAccumulation_TTLEviction(t *testing.T) {
	tr := newTestTransport()
	defer close(tr.stopCleanup)

	// Inject a stale entry directly.
	stale := &pendingChunks{
		chunks:    make(map[int][]byte),
		total:     2,
		createdAt: time.Time{}, // year 1 — always expired
	}
	tr.chunkStore.Store("stale", stale)

	// A fresh entry that must NOT be evicted.
	fresh := &pendingChunks{
		chunks:    make(map[int][]byte),
		total:     2,
		createdAt: time.Now(),
	}
	tr.chunkStore.Store("fresh", fresh)

	tr.sweepExpired()

	if _, ok := tr.chunkStore.Load("stale"); ok {
		t.Error("stale entry was not evicted by sweepExpired")
	}
	if _, ok := tr.chunkStore.Load("fresh"); !ok {
		t.Error("fresh entry was incorrectly evicted by sweepExpired")
	}
}
