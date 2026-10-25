package task

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/fmotalleb/go-tools/env"
)

var prependDateTime = sync.OnceValue(func() bool {
	return env.BoolOr("CRONTAB_ERR_LOG_DATE_TIME", false)
})

// maxPrefixLen is a Grow() sizing hint: the largest a per-line prefix can be
// (RFC3339 timestamp + space, worst case 25 bytes for a non-"Z" numeric
// offset) plus the id and " | " suffix are added on top per writer.
const maxPrefixLen = len(time.RFC3339) + 1

// prefixWriter prefixes every new line with "[<RFC3339 time> ]<id> | " before
// forwarding to dst. The timestamp, when enabled, is computed fresh for each
// line as it starts — not once at construction — so long-running lines are
// stamped with the time they actually began.
// It streams line-by-line without buffering the whole output in memory long-term;
// a single reusable scratch buffer is used to assemble each Write call's output so
// that only one underlying dst.Write happens per call, regardless of how many lines
// p contains.
type prefixWriter struct {
	mu        sync.Mutex
	dst       io.Writer
	id        []byte
	lineStart bool
	scratch   bytes.Buffer // reused across Write calls; guarded by mu
}

// NewPrefixWriter returns an io.Writer that prefixes every line with the
// given id (and, if CRONTAB_ERR_LOG_DATE_TIME is set, the current time at
// the moment that line starts).
func NewPrefixWriter(dst io.Writer, id string) io.Writer {
	return &prefixWriter{
		dst:       dst,
		id:        []byte(id),
		lineStart: true,
	}
}

func (w *prefixWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	w.scratch.Reset()
	nlCount := bytes.Count(p, []byte{'\n'})
	perLinePrefixLen := len(w.id) + 3 // id + " | "
	if prependDateTime() {
		perLinePrefixLen += maxPrefixLen
	}
	// Grow up-front: worst case is a prefix on every line, plus one forced
	// trailing newline byte (see below).
	w.scratch.Grow(len(p) + (nlCount+1)*perLinePrefixLen + 1)

	start := 0
	for i, b := range p {
		if b == '\n' {
			w.appendChunk(p[start : i+1])
			start = i + 1
		}
	}
	if start < len(p) {
		w.appendChunk(p[start:])
		// p didn't end on a newline. Force one so this Write call always
		// terminates its own line: the next Write starts a fresh, prefixed
		// line instead of silently continuing this one.
		w.scratch.WriteByte('\n')
		w.lineStart = true
	}

	// Single write to dst for the whole (possibly multi-line) input, instead of
	// one write per line.
	if _, err := w.dst.Write(w.scratch.Bytes()); err != nil {
		return 0, err
	}
	return len(p), nil
}

// appendChunk appends chunk to the scratch buffer, adding the prefix first if
// this chunk begins a new line. The prefix (including the timestamp, if
// enabled) is computed at the moment the line starts, so it reflects when
// that specific line began rather than when the writer was constructed.
// It never returns an error: writes to a bytes.Buffer only fail on
// out-of-memory, which panics rather than erroring.
func (w *prefixWriter) appendChunk(chunk []byte) {
	if len(chunk) == 0 {
		return
	}
	if w.lineStart {
		if prependDateTime() {
			w.scratch.WriteString(time.Now().Format(time.RFC3339))
			w.scratch.WriteByte(' ')
		}
		w.scratch.Write(w.id)
		w.scratch.WriteString(" | ")
	}
	w.scratch.Write(chunk)
	w.lineStart = chunk[len(chunk)-1] == '\n'
}

// shortHash returns a compact hex digest of the given string, enough to tell runs apart.
func shortHash(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:4])
}

// newExecutionID returns a random identifier unique to each task execution.
func newExecutionID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%x", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}
