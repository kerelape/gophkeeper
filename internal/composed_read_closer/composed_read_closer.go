// Package composedreadcloser provides ComposedReadCloser,
// which can be use to compbine an io.Reader and an io.Closer
// together in a single object.
package composedreadcloser

import "io"

// ComposedReadCloser is a composed io.ReadCloser.
type ComposedReadCloser struct {
	Reader io.Reader
	Closer io.Closer
}

// Read implements io.Reader.
func (rc *ComposedReadCloser) Read(p []byte) (int, error) {
	return rc.Reader.Read(p)
}

// Close implements io.ReadCloser.
func (rc *ComposedReadCloser) Close() error {
	return rc.Closer.Close()
}
