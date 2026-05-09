package application

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

// newTestTransport builds an HTTPTransport with a processBodyFn test hook so
// that unit tests can inspect assembled payloads without needing a real
// MessageProcessor or running HTTP server.
func newTestTransport(fn func(http.ResponseWriter, *http.Request, []byte)) *HTTPTransport {
	tr := NewHTTPTransport()
	tr.processBodyFn = fn
	tr.stopCleanup = make(chan struct{})
	go tr.cleanupChunks()
	return tr
}

func makeChunkReq(chunkID string, index, total int, body []byte) *http.Request {
	req := httptest.NewRequest(http.MethodPost, "/wails/runtime", bytes.NewReader(body))
	req.Header.Set(chunkIDHeader, chunkID)
	req.Header.Set(chunkIndexHeader, fmt.Sprintf("%d", index))
	req.Header.Set(chunkTotalHeader, fmt.Sprintf("%d", total))
	return req
}

// ————————————————————————————————————————————————————————
// Validation tests
// ————————————————————————————————————————————————————————

func TestChunkValidation_IndexNegative(t *testing.T) {
	tr := newTestTransport(nil)
	defer tr.Stop()
	req := makeChunkReq("id1", 0, 3, []byte("x"))
	req.Header.Set(chunkIndexHeader, "-1")
	rr := httptest.NewRecorder()
	tr.handleChunkedRequest(rr, req, "id1")
	if rr.Code == http.StatusOK {
		t.Fatal("expected non-200 for negative index")
	}
}

func TestChunkValidation_IndexGeTotal(t *testing.T) {
	tr := newTestTransport(nil)
	defer tr.Stop()
	// index == total is invalid (0-based, must be < total)
	req := makeChunkReq("id2", 3, 3, []byte("x"))
	rr := httptest.NewRecorder()
	tr.handleChunkedRequest(rr, req, "id2")
	if rr.Code == http.StatusOK {
		t.Fatal("expected non-200 for index >= total")
	}
}

func TestChunkValidation_TotalZero(t *testing.T) {
	tr := newTestTransport(nil)
	defer tr.Stop()
	req := makeChunkReq("id3", 0, 1, []byte("x"))
	req.Header.Set(chunkTotalHeader, "0")
	rr := httptest.NewRecorder()
	tr.handleChunkedRequest(rr, req, "id3")
	if rr.Code == http.StatusOK {
		t.Fatal("expected non-200 for total=0")
	}
}

func TestChunkValidation_TotalExceedsMax(t *testing.T) {
	tr := newTestTransport(nil)
	defer tr.Stop()
	req := makeChunkReq("id4", 0, maxChunkTotal+1, []byte("x"))
	rr := httptest.NewRecorder()
	tr.handleChunkedRequest(rr, req, "id4")
	if rr.Code == http.StatusOK {
		t.Fatalf("expected non-200 for total exceeding maxChunkTotal (%d)", maxChunkTotal)
	}
}

func TestChunkValidation_TotalMismatch(t *testing.T) {
	tr := newTestTransport(nil)
	defer tr.Stop()

	// First chunk establishes total=3.
	rr1 := httptest.NewRecorder()
	tr.handleChunkedRequest(rr1, makeChunkReq("id5", 0, 3, []byte("aaa")), "id5")
	if rr1.Code != http.StatusOK {
		t.Fatalf("first chunk rejected unexpectedly: %d", rr1.Code)
	}

	// Second chunk claims total=2 — mismatch.
	req2 := makeChunkReq("id5", 1, 2, []byte("bbb"))
	rr2 := httptest.NewRecorder()
	tr.handleChunkedRequest(rr2, req2, "id5")
	if rr2.Code == http.StatusOK {
		t.Fatal("expected non-200 for total mismatch on second chunk")
	}
}

func TestChunkValidation_InvalidIndexString(t *testing.T) {
	for _, bad := range []string{"abc", "", "1.5"} {
		tr := newTestTransport(nil)
		defer tr.Stop()
		req := makeChunkReq("id-bad", 0, 3, []byte("x"))
		req.Header.Set(chunkIndexHeader, bad)
		rr := httptest.NewRecorder()
		tr.handleChunkedRequest(rr, req, "id-bad")
		if rr.Code == http.StatusOK {
			t.Errorf("expected non-200 for index %q", bad)
		}
	}
}

func TestChunkValidation_InvalidTotalString(t *testing.T) {
	for _, bad := range []string{"xyz", "", "-5"} {
		tr := newTestTransport(nil)
		defer tr.Stop()
		req := makeChunkReq("id-bad2", 0, 3, []byte("x"))
		req.Header.Set(chunkTotalHeader, bad)
		rr := httptest.NewRecorder()
		tr.handleChunkedRequest(rr, req, "id-bad2")
		if rr.Code == http.StatusOK {
			t.Errorf("expected non-200 for total %q", bad)
		}
	}
}

// ————————————————————————————————————————————————————————
// Assembly correctness
// ————————————————————————————————————————————————————————

func TestChunkAssembly_Correct(t *testing.T) {
	cases := []struct {
		name        string
		payloadSize int
		chunkSize   int
	}{
		{"1KB_1chunk", 1024, 512 * 1024},
		{"512KB_2chunks", 512 * 1024, 256 * 1024},
		{"1MB_2chunks", 1 * 1024 * 1024, 512 * 1024},
		{"2MB_4chunks", 2 * 1024 * 1024, 512 * 1024},
		{"3MB_6chunks", 3 * 1024 * 1024, 512 * 1024},
		{"5MB_10chunks", 5 * 1024 * 1024, 512 * 1024},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			payload := bytes.Repeat([]byte("A"), tc.payloadSize)
			total := (tc.payloadSize + tc.chunkSize - 1) / tc.chunkSize
			chunkID := "assembly-" + tc.name

			var assembled []byte
			var once sync.Once
			tr := newTestTransport(func(_ http.ResponseWriter, _ *http.Request, body []byte) {
				once.Do(func() { assembled = body })
			})
			defer tr.Stop()

			for i := 0; i < total; i++ {
				start := i * tc.chunkSize
				end := start + tc.chunkSize
				if end > len(payload) {
					end = len(payload)
				}
				rr := httptest.NewRecorder()
				tr.handleChunkedRequest(rr, makeChunkReq(chunkID, i, total, payload[start:end]), chunkID)
				if i < total-1 && rr.Code != http.StatusOK {
					t.Fatalf("chunk %d/%d rejected: %d", i, total, rr.Code)
				}
			}

			if !bytes.Equal(assembled, payload) {
				t.Fatalf("assembly mismatch: got %d bytes, want %d", len(assembled), len(payload))
			}
		})
	}
}

func TestChunkAssembly_IntermediateChunksReturn200(t *testing.T) {
	tr := newTestTransport(func(rw http.ResponseWriter, _ *http.Request, _ []byte) {
		rw.WriteHeader(http.StatusOK)
	})
	defer tr.Stop()
	chunkID := "intermediate-200"

	for i := 0; i < 4; i++ {
		rr := httptest.NewRecorder()
		tr.handleChunkedRequest(rr, makeChunkReq(chunkID, i, 5, []byte("data")), chunkID)
		if rr.Code != http.StatusOK {
			t.Fatalf("intermediate chunk %d returned %d, want 200", i, rr.Code)
		}
	}
}

func TestChunkAssembly_ChunkStoreCleared(t *testing.T) {
	chunkID := "store-clear-test"
	tr := newTestTransport(func(_ http.ResponseWriter, _ *http.Request, _ []byte) {})
	defer tr.Stop()

	tr.handleChunkedRequest(httptest.NewRecorder(), makeChunkReq(chunkID, 0, 1, []byte("hello")), chunkID)

	if _, loaded := tr.chunkStore.Load(chunkID); loaded {
		t.Fatal("chunkStore entry not deleted after assembly")
	}
}

func TestChunkAssembly_DuplicateChunkIgnored(t *testing.T) {
	chunkID := "dup-test"
	chunk0 := []byte("first")
	chunk1 := []byte("second")

	var assembled []byte
	tr := newTestTransport(func(_ http.ResponseWriter, _ *http.Request, body []byte) {
		assembled = body
	})
	defer tr.Stop()

	// Send chunk 0 twice — second should be a no-op.
	tr.handleChunkedRequest(httptest.NewRecorder(), makeChunkReq(chunkID, 0, 2, chunk0), chunkID)
	tr.handleChunkedRequest(httptest.NewRecorder(), makeChunkReq(chunkID, 0, 2, []byte("DUPLICATE")), chunkID)
	tr.handleChunkedRequest(httptest.NewRecorder(), makeChunkReq(chunkID, 1, 2, chunk1), chunkID)

	expected := append(chunk0, chunk1...)
	if !bytes.Equal(assembled, expected) {
		t.Fatalf("got %q, want %q", assembled, expected)
	}
}

func TestChunkAssembly_CompletedFlagPreventDoubleProcess(t *testing.T) {
	chunkID := "double-process"
	payload := []byte("data")

	callCount := 0
	var mu sync.Mutex
	tr := newTestTransport(func(rw http.ResponseWriter, _ *http.Request, _ []byte) {
		mu.Lock()
		callCount++
		mu.Unlock()
		rw.WriteHeader(http.StatusOK)
	})
	defer tr.Stop()

	// Send the only chunk, completing the transfer.
	tr.handleChunkedRequest(httptest.NewRecorder(), makeChunkReq(chunkID, 0, 1, payload), chunkID)

	// Inject a pre-completed pc entry to simulate a concurrent late arrival.
	cloneID := chunkID + "-clone"
	pc := &pendingChunks{
		chunks:        map[int][]byte{0: payload},
		total:         1,
		aggregateSize: len(payload),
		completed:     true,
		createdAt:     time.Now(),
		lastSeen:      time.Now(),
	}
	tr.chunkStore.Store(cloneID, pc)
	rr := httptest.NewRecorder()
	tr.handleChunkedRequest(rr, makeChunkReq(cloneID, 0, 1, payload), cloneID)

	mu.Lock()
	defer mu.Unlock()
	if callCount != 1 {
		t.Fatalf("processBody called %d times, want 1", callCount)
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("late duplicate returned %d, want 200", rr.Code)
	}
}

// ————————————————————————————————————————————————————————
// TTL eviction
// ————————————————————————————————————————————————————————

func TestChunkTTL_Eviction(t *testing.T) {
	tr := NewHTTPTransport()
	tr.stopCleanup = make(chan struct{})
	defer tr.Stop()

	chunkID := "ttl-expired"
	past := time.Now().Add(-(chunkTTL + time.Second))
	tr.chunkStore.Store(chunkID, &pendingChunks{
		chunks:    make(map[int][]byte),
		total:     2,
		createdAt: past,
		lastSeen:  past,
	})

	tr.sweepExpired()

	if _, loaded := tr.chunkStore.Load(chunkID); loaded {
		t.Fatal("expired entry not evicted by sweepExpired")
	}
}

func TestChunkTTL_ActiveTransferNotEvicted(t *testing.T) {
	tr := NewHTTPTransport()
	tr.stopCleanup = make(chan struct{})
	defer tr.Stop()

	chunkID := "ttl-active"
	tr.chunkStore.Store(chunkID, &pendingChunks{
		chunks:    make(map[int][]byte),
		total:     2,
		createdAt: time.Now().Add(-60 * time.Second),
		lastSeen:  time.Now(), // recently active
	})

	tr.sweepExpired()

	if _, loaded := tr.chunkStore.Load(chunkID); !loaded {
		t.Fatal("active entry incorrectly evicted")
	}
}
