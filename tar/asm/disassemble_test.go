package asm

import (
	"archive/tar"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/vbatts/tar-split/tar/storage"
)

// This test failing causes the binary to crash due to memory overcommitment.
func TestLargeJunkPadding(t *testing.T) {
	pR, pW := io.Pipe()

	// Write a normal tar file into the pipe and then load it full of junk
	// bytes as padding. We have to do this in a goroutine because we can't
	// store 20GB of junk in-memory.
	go func() {
		// Empty archive.
		tw := tar.NewWriter(pW)
		if err := tw.Close(); err != nil {
			pW.CloseWithError(err)
			return
		}

		// Write junk.
		const (
			junkChunkSize = 64 * 1024 * 1024
			junkChunkNum  = 20 * 16
		)
		devZero, err := os.Open("/dev/zero")
		if err != nil {
			pW.CloseWithError(err)
			return
		}
		defer devZero.Close()
		for i := 0; i < junkChunkNum; i++ {
			if i%32 == 0 {
				fmt.Fprintf(os.Stderr, "[TestLargeJunkPadding] junk chunk #%d/#%d\n", i, junkChunkNum)
			}
			if _, err := io.CopyN(pW, devZero, junkChunkSize); err != nil {
				pW.CloseWithError(err)
				return
			}
		}

		fmt.Fprintln(os.Stderr, "[TestLargeJunkPadding] junk chunk finished")
		pW.Close()
	}()

	// Disassemble our junk file.
	nilPacker := storage.NewJSONPacker(io.Discard)
	rdr, err := NewInputTarStream(pR, nilPacker, nil)
	if err != nil {
		t.Fatal(err)
	}

	// Copy the entire rdr.
	_, err = io.Copy(io.Discard, rdr)
	if err != nil {
		t.Fatal(err)
	}

	// At this point, if we haven't crashed then we are not vulnerable to
	// CVE-2017-14992.
}

// Mocked Packer storing entries and returning an error on demand.
type recordingPacker struct {
	mu      sync.Mutex
	entries []storage.Entry
	errAt   int
	err     error
	callNum int
}

func (p *recordingPacker) AddEntry(e storage.Entry) (int, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	// Return an aritifical error if we are instructed to do so.
	p.callNum++
	if p.errAt > 0 && p.callNum == p.errAt {
		if p.err == nil {
			p.err = errors.New("packer error")
		}
		return 0, p.err
	}

	// Copy payload because callers may reuse buffers.
	cp := e
	if e.Payload != nil {
		cp.Payload = append([]byte(nil), e.Payload...)
	}
	p.entries = append(p.entries, cp)
	return len(cp.Payload), nil
}

func (p *recordingPacker) snapshot() []storage.Entry {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]storage.Entry, len(p.entries))
	copy(out, p.entries)
	return out
}

// Mocked FilePutter
type recordingFilePutter struct {
	mu   sync.Mutex
	puts []string
}

func (fp *recordingFilePutter) Put(name string, r io.Reader) (int64, []byte, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return 0, nil, err
	}
	fp.mu.Lock()
	fp.puts = append(fp.puts, name)
	fp.mu.Unlock()

	// Return a deterministic "checksum" based on content length.
	csum := []byte(fmt.Sprintf("len=%d", len(b)))
	return int64(len(b)), csum, nil
}

// Helper function to generate the tar with optional extra padding.
func makeTarWithExtraPadding(t *testing.T, name string, content []byte, extraPadding int) []byte {
	t.Helper()

	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	hdr := &tar.Header{
		Name: name,
		Mode: 0o644,
		Size: int64(len(content)),
	}
	if err := tw.WriteHeader(hdr); err != nil {
		t.Fatalf("WriteHeader: %v", err)
	}
	if _, err := tw.Write(content); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("Close tar writer: %v", err)
	}

	out := buf.Bytes()
	if extraPadding > 0 {
		out = append(append([]byte(nil), out...), make([]byte, extraPadding)...)
	}
	return out
}

// Helper function to wait until "done" for specific time.
func waitDone(t *testing.T, done <-chan error) error {
	t.Helper()
	select {
	case err := <-done:
		return err
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for done")
		return errors.New("timeout")
	}
}

// closableBlockingReader simulates an io.Reader that can be "closed" while a Read is
// blocked.
//
// Behavior:
// - It serves bytes from data.
// - After it has served at least blockAfter bytes, the next Read blocks until either:
//   - Unblock() is called, or
//   - Close() is called (which also unblocks) and subsequent reads fail with errUnderlyingClosed.
type closableBlockingReader struct {
	data       []byte
	pos        int
	blockAfter int

	closed atomic.Bool

	blockOnce sync.Once
	blockCh   chan struct{} // used to block/unblock reads
	blockedCh chan struct{} // closed when we start blocking
}

var errUnderlyingClosed = errors.New("underlying reader closed")

func newClosableBlockingReader(data []byte, blockAfter int) *closableBlockingReader {
	return &closableBlockingReader{
		data:       data,
		blockAfter: blockAfter,
		blockCh:    make(chan struct{}),
		blockedCh:  make(chan struct{}),
	}
}

func (r *closableBlockingReader) Read(p []byte) (int, error) {
	if r.closed.Load() {
		return 0, errUnderlyingClosed
	}
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}

	// If we've reached the point where we should block, block before producing
	// more data (simulates "reader got closed while goroutine is still running").
	if r.pos >= r.blockAfter {
		r.blockOnce.Do(func() { close(r.blockedCh) })
		<-r.blockCh
		if r.closed.Load() {
			return 0, errUnderlyingClosed
		}
	}

	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}

func (r *closableBlockingReader) Close() error {
	r.closed.Store(true)
	// ensure blocked goroutine wakes up
	select {
	case <-r.blockCh:
		// already closed/unblocked
	default:
		close(r.blockCh)
	}
	// signal blocked state even if we closed early
	r.blockOnce.Do(func() { close(r.blockedCh) })
	return nil
}

func (r *closableBlockingReader) Unblock() {
	select {
	case <-r.blockCh:
	default:
		close(r.blockCh)
	}
}

// Test that NewInputTarStreamWithDone signals done when we read everything.
func TestNewInputTarStreamWithDone(t *testing.T) {
	input := makeTarWithExtraPadding(t, "file.txt", []byte("hello"), 4096)

	p := &recordingPacker{}
	fp := &recordingFilePutter{}

	payload, done, err := NewInputTarStreamWithDone(bytes.NewReader(input), p, fp)
	if err != nil {
		t.Fatalf("NewInputTarStreamWithDone: %v", err)
	}
	defer payload.Close()

	got, rerr := io.ReadAll(payload)
	if rerr != nil {
		t.Fatalf("ReadAll(payload): %v", rerr)
	}
	if !bytes.Equal(got, input) {
		t.Fatalf("payload bytes differ: got=%d bytes, want=%d bytes", len(got), len(input))
	}

	if derr := waitDone(t, done); derr != nil {
		t.Fatalf("done returned error: %v", derr)
	}

	entries := p.snapshot()
	if len(entries) == 0 {
		t.Fatalf("expected entries to be recorded")
	}

	var (
		foundFile    bool
		foundSegment bool
	)
	for _, e := range entries {
		switch e.Type {
		case storage.FileType:
			foundFile = true
			// We set size to len("hello")
			if e.Size != int64(len("hello")) {
				t.Fatalf("file entry size=%d, want=%d", e.Size, len("hello"))
			}
		case storage.SegmentType:
			if len(e.Payload) > 0 {
				foundSegment = true
			}
		}
	}
	if !foundFile {
		t.Fatalf("expected at least one FileType entry")
	}
	if !foundSegment {
		t.Fatalf("expected at least one SegmentType entry with payload")
	}
}

// Test that NewInputTarStreamWithDone works when underlying reader is closed while
// the NewInputTarStreamWithDone go-routine still runs.
func TestNewInputTarStreamWithDonUnderlyingClosed(t *testing.T) {
	// Make a tar stream that is large enough that parsing won't finish in one tiny read.
	input := makeTarWithExtraPadding(t, "file.txt", bytes.Repeat([]byte("A"), 64*1024), 0)

	// Block the underlying reader after it has produced some bytes.
	// This ensures the tar-split goroutine will be mid-flight and will need more data.
	under := newClosableBlockingReader(input, 4096)

	p := &recordingPacker{}
	fp := storage.NewDiscardFilePutter()

	payload, done, err := NewInputTarStreamWithDone(under, p, fp)
	if err != nil {
		t.Fatalf("NewInputTarStreamWithDone: %v", err)
	}
	defer payload.Close()

	// Start draining payload in a separate goroutine so the internal goroutine is forced to read.
	readErrCh := make(chan error, 1)
	go func() {
		_, rerr := io.ReadAll(payload)
		readErrCh <- rerr
	}()

	// Wait until the underlying reader starts blocking (i.e., internal goroutine progressed
	// far enough to need more bytes).
	select {
	case <-under.blockedCh:
		// good
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for underlying reader to enter blocked state")
	}

	// Now "close" the underlying reader while the tar-split goroutine is still running.
	under.Close()

	// The tar-split goroutine should treat this as a non-EOF error, call fail(err),
	// CloseWithError on the pipe, and signal done with the same error.
	derr := waitDone(t, done)
	if derr == nil {
		t.Fatalf("expected done error, got nil")
	}
	if !errors.Is(derr, errUnderlyingClosed) {
		t.Fatalf("done error=%v, want errors.Is(..., errUnderlyingClosed)=true", derr)
	}

	// The consumer side should also observe an error (from the pipe).
	select {
	case rerr := <-readErrCh:
		if rerr == nil {
			t.Fatalf("expected reader error, got nil")
		}
		if !errors.Is(rerr, errUnderlyingClosed) {
			t.Fatalf("reader error=%v, want errors.Is(..., errUnderlyingClosed)=true", rerr)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timeout waiting for payload read to finish")
	}
}

// Test that if the caller closes the returned reader without draining it,
// the background goroutine terminates promptly and the done channel is signaled.
func TestNewInputTarStreamWithDoneEarlyClose(t *testing.T) {
	input := makeTarWithExtraPadding(t, "file.txt", []byte("hello"), 2048)

	p := &recordingPacker{}
	fp := &recordingFilePutter{}

	payload, done, err := NewInputTarStreamWithDone(bytes.NewReader(input), p, fp)
	if err != nil {
		t.Fatalf("NewInputTarStreamWithDone: %v", err)
	}

	// Close immediately without draining: this should abort the pipe and allow
	// the goroutine to exit (rather than hanging).
	if err := payload.Close(); err != nil {
		t.Fatalf("payload.Close(): %v", err)
	}

	// The goroutine should terminate and signal done (likely with a non-nil error
	// due to the induced abort).
	derr := waitDone(t, done)
	if !errors.Is(derr, io.ErrClosedPipe) {
		t.Fatalf("done error=%v, want io.ErrClosedPipe", derr)
	}
}

// Test that Packer error propagates to waitDone().
func TestNewInputTarStreamWithDonePackerError(t *testing.T) {
	input := makeTarWithExtraPadding(t, "file.txt", []byte("hello"), 0)

	packerErr := errors.New("boom")
	p := &recordingPacker{errAt: 2, err: packerErr} // fail early during AddEntry
	fp := &recordingFilePutter{}

	payload, done, err := NewInputTarStreamWithDone(bytes.NewReader(input), p, fp)
	if err != nil {
		t.Fatalf("NewInputTarStreamWithDone: %v", err)
	}
	defer payload.Close()

	// Reading should eventually return the packer error via CloseWithError.
	_, rerr := io.ReadAll(payload)
	if rerr == nil {
		t.Fatalf("expected reader error, got nil")
	}
	// The error returned from an io.PipeReader may wrap; check with errors.Is.
	if !errors.Is(rerr, packerErr) {
		t.Fatalf("reader error=%v, want errors.Is(...,%v)=true", rerr, packerErr)
	}

	derr := waitDone(t, done)
	if derr == nil {
		t.Fatalf("expected done error, got nil")
	}
	if !errors.Is(derr, packerErr) {
		t.Fatalf("done error=%v, want errors.Is(...,%v)=true", derr, packerErr)
	}
}
