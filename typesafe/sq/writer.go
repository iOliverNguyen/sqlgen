package sq

import (
	"strconv"
	"unsafe"

	"github.com/ng-vu/sqlgen/core"
)

var _ core.SQLWriter = &Writer{}

type Writer struct {
	c int64

	opts   core.Opts
	quote  byte
	marker byte

	buf  []byte
	args []interface{}
	scan []interface{}
}

func NewWriter(opts core.Opts, quote byte, marker byte, size int) *Writer {
	return &Writer{
		quote:  quote,
		marker: marker,
		opts:   opts,
		buf:    make([]byte, 0, size),
		args:   make([]interface{}, 0, 64),
	}
}

func (w *Writer) sqlwriter() {}

func (w *Writer) Opts() core.Opts {
	return w.opts
}

// TrimLast removes n last bytes written
func (w *Writer) TrimLast(n int) {
	w.buf = w.buf[:len(w.buf)-n]
}

func (w *Writer) Len() int {
	return len(w.buf)
}

func (w *Writer) WriteArg(arg interface{}) {
	w.args = append(w.args, arg)
}

func (w *Writer) WriteArgs(args []interface{}) {
	w.args = append(w.args, args...)
}

func (w *Writer) WriteScanArg(arg interface{}) {
	w.scan = append(w.scan, arg)
}

func (w *Writer) WriteScanArgs(args []interface{}) {
	w.scan = append(w.scan, args...)
}

func (w *Writer) WriteMarker() {
	if w.marker == '?' {
		w.buf = appendQuestionMarker(w.buf)
	} else {
		w.buf = appendDollarMarker(w.buf, &w.c)
	}
}

func (w *Writer) WriteMarkers(n int) {
	if w.marker == '?' {
		w.buf = appendQuestionMarkers(w.buf, n)
	} else {
		w.buf = appendDollarMarkers(w.buf, &w.c, n)
	}
}

func (w *Writer) WriteQuery(query []byte) {
	w.buf = appendAndReplace(w.buf, &w.c, w.quote, w.marker, unsafeBytesToString(query), "")
}

func (w *Writer) WriteQueryString(query string) {
	w.buf = appendAndReplace(w.buf, &w.c, w.quote, w.marker, query, "")
}

func (w *Writer) WriteQueryStringWithPrefix(prefix, query string) {
	w.buf = appendAndReplace(w.buf, &w.c, w.quote, w.marker, query, prefix)
}

func (w *Writer) WritePrefixedName(schema, name string) {
	if schema != "" {
		w.buf = append(w.buf, schema...)
		w.buf = append(w.buf, '.')
	}
	w.buf = append(w.buf, w.quote)
	w.buf = append(w.buf, name...)
	w.buf = append(w.buf, w.quote)
}

func (w *Writer) WriteName(name string) {
	w.buf = append(w.buf, w.quote)
	w.buf = append(w.buf, name...)
	w.buf = append(w.buf, w.quote)
}

func (w *Writer) WriteQueryName(name string) {
	if shouldQuote(name) {
		w.buf = append(w.buf, w.quote)
		w.buf = append(w.buf, name...)
		w.buf = append(w.buf, w.quote)
	} else {
		w.WriteQueryString(name)
	}
}

func (w *Writer) WriteByte(b byte) {
	w.buf = append(w.buf, b)
}

func (w *Writer) WriteRaw(query []byte) {
	w.buf = append(w.buf, query...)
}

func (w *Writer) WriteRawString(query string) {
	w.buf = append(w.buf, query...)
}

func (w *Writer) String() string {
	return unsafeBytesToString(w.buf)
}

func (w *Writer) Args() []interface{} {
	return w.args
}

func (w *Writer) ScanArgs() []interface{} {
	return w.scan
}

func appendQuestionMarker(b []byte) []byte {
	b = append(b, '?')
	return b
}

func appendQuestionMarkers(b []byte, n int) []byte {
	b = append(b, '?')
	for i := 1; i < n; i++ {
		b = append(b, ",?"...)
	}
	return b
}

func appendDollarMarker(b []byte, c *int64) []byte {
	*c++
	b = append(b, '$')
	b = strconv.AppendInt(b, *c, 10)
	return b
}

func appendDollarMarkers(b []byte, c *int64, n int) []byte {
	*c++
	b = append(b, '$')
	b = strconv.AppendInt(b, *c, 10)
	for i := 1; i < n; i++ {
		*c++
		b = append(b, ",$"...)
		b = strconv.AppendInt(b, *c, 10)
	}
	return b
}

func appendAndReplace(b []byte, c *int64, quote, marker byte, query string, schema string) []byte {
	last := byte(0)
	idx := 0
	for i := 0; i < len(query); i++ {
		ch := query[i]
		switch {
		case ch == '"' && ch != quote:
			b = append(b, query[idx:i]...)
			b = append(b, quote)
			idx = i + 1
		case ch == '?' && ch != marker:
			*c++
			b = append(b, query[idx:i]...)
			b = append(b, marker)
			b = strconv.AppendInt(b, *c, 10)
			idx = i + 1
		case ch == '.' && last == '$':
			b = append(b, query[idx:i-1]...)
			if schema != "" {
				b = append(b, schema...)
				b = append(b, '.')
			}
			idx = i + 1
		}
		last = ch
	}
	if idx < len(query) {
		b = append(b, query[idx:]...)
	}
	return b
}

func shouldQuote(s string) bool {
	if s == "" {
		panic("sqlgen: empty name")
	}
	c := s[0]
	if !(c == '_' ||
		c >= 'a' && c <= 'z' ||
		c >= 'A' && c <= 'Z') {
		return false
	}
	for i := range s {
		c := s[i]
		if !(c == '_' ||
			c >= 'a' && c <= 'z' ||
			c >= 'A' && c <= 'Z' ||
			c >= '0' && c <= '9') {
			return false
		}
	}
	return true
}

//go:nosplit
func unsafeBytesToString(b []byte) string {
	return *(*string)(unsafe.Pointer(&b))
}
