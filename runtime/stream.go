package runtime

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"
)

const (
	STREAM_FLAG_READ   = os.O_RDONLY
	STREAM_FLAG_WRITE  = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	STREAM_FLAG_RW     = os.O_RDWR | os.O_CREATE
	STREAM_FLAG_APPEND = os.O_WRONLY | os.O_APPEND | os.O_CREATE
)

type Stream interface {
	io.Reader
	io.Writer
	io.Closer
}

type Buffer struct {
	buf      *bytes.Buffer
	readonly bool
}

func (s *Buffer) Close() error {
	if s.buf == nil {
		return fmt.Errorf("cannot close closed stream")
	}
	s.buf = nil
	return nil
}

func (s *Buffer) Read(p []byte) (n int, err error) {
	if s.buf == nil {
		return 0, fmt.Errorf("bad file descriptor, cannot read from closed stream")
	}
	return s.buf.Read(p)
}

func (s *Buffer) Write(p []byte) (n int, err error) {
	if s.buf == nil {
		return 0, fmt.Errorf("bad file descriptor, cannot write to closed stream")
	}
	if s.readonly {
		return 0, fmt.Errorf("bad file descriptor, cannot write to read-only stream")
	}
	return s.buf.Write(p)
}

func (s *Buffer) String(trim_leading_newline bool) string {
	v := s.buf.String()
	if trim_leading_newline {
		return strings.TrimRight(v, "\n")
	}
	return v
}

func NewBuffer(s string, readonly bool) *Buffer {
	return &Buffer{
		buf:      bytes.NewBufferString(s),
		readonly: readonly,
	}
}

type proxyStream struct {
	original Stream
	closed   bool
}

func (s *proxyStream) Close() error {
	if s.closed {
		return fmt.Errorf("cannot close closed stream")
	}
	s.closed = true
	return nil
}

func (s *proxyStream) Read(p []byte) (n int, err error) { return 0, nil }

func (s *proxyStream) Write(p []byte) (n int, err error) { return 0, nil }
func (s *proxyStream) getOriginal() (Stream, error) {
	if s.closed {
		return nil, fmt.Errorf("file descriptor is closed")
	}

	if o, ok := s.original.(*proxyStream); ok {
		return o.getOriginal()
	}

	return s.original, nil
}

type StreamManager struct {
	mappings map[string]Stream
	proxied  []Stream
}

func (sm *StreamManager) OpenStream(name string, flag int) (Stream, error) {
	switch name {
	default:
		return os.OpenFile(name, flag, 0644)
	}
}

func (sm *StreamManager) Add(fd string, stream Stream, proxy bool) {
	// If this stream is already open, we need to close it. otherwise, Its handler will be lost and leak.
	// This is related to pipelines in particular. when instantiating a new pipeline, we add its ends to the FDT. but if
	// a redirection happened afterwards, it will cause the pipline handler to be lost and kept open.
	if sm.mappings[fd] != nil {
		sm.mappings[fd].Close()
	}

	if proxy {
		sm.proxied = append(sm.proxied, stream)
		stream = &proxyStream{original: stream}
	}

	sm.mappings[fd] = stream
}

func (sm *StreamManager) Get(fd string) (Stream, error) {
	stream, ok := sm.mappings[fd]
	if !ok {
		return nil, fmt.Errorf("file descriptor %q is not open", fd)
	}

	if p, ok := stream.(*proxyStream); ok {
		if o, err := p.getOriginal(); err != nil {
			return nil, fmt.Errorf("bad file descriptor %q, %w", fd, err)
		} else {
			return o, nil
		}
	}

	return stream, nil
}

func (sm *StreamManager) Duplicate(newfd, oldfd string) error {
	if stream, ok := sm.mappings[oldfd]; !ok {
		return fmt.Errorf("trying to duplicate bad file descriptor: %s", oldfd)
	} else {
		// when trying to duplicate a file descriptor to it self (eg: 3>&3 ), we just return.
		if newfd == oldfd {
			return nil
		}

		// If the new fd is already open, we need to close it. otherwise, Its handler will be lost and leak. and remain open forever.
		// for example: "3<file.txt 3<&0", we don't explicitly close 3. Thus, it is going to remain open forever, unless we implicitly close it here.
		if sm.mappings[newfd] != nil {
			sm.mappings[newfd].Close()
		}

		switch stream := stream.(type) {
		case *Buffer:
			newbuf := &bytes.Buffer{}
			_, err := io.Copy(newbuf, stream)
			if err != nil {
				return fmt.Errorf("failed to duplicate file descriptor '%s', %w", oldfd, err)
			}
			sm.mappings[newfd] = &Buffer{buf: newbuf, readonly: stream.readonly}
		case *os.File:
			dupFd, err := syscall.Dup(int(stream.Fd()))
			if err != nil {
				return fmt.Errorf("failed to duplicate file descriptor '%s', %w", oldfd, err)
			}
			sm.mappings[newfd] = os.NewFile(uintptr(dupFd), stream.Name())
		case *proxyStream:
			sm.mappings[newfd] = &proxyStream{
				original: stream.original,
			}
		default:
			panic(fmt.Sprintf("failed to clone (%s), unhandled stream type: %T", oldfd, stream))
		}

		return nil
	}
}

func (sm *StreamManager) Close(fd string) error {
	if stream, ok := sm.mappings[fd]; !ok {
		return fmt.Errorf("trying to close bad file descriptor: %s", fd)
	} else {
		return stream.Close()
	}
}

func (sm *StreamManager) Destroy() {
	for _, stream := range sm.proxied {
		stream.Close()
	}
	for _, stream := range sm.mappings {
		stream.Close()
	}
}

func (sm *StreamManager) Clone() *StreamManager {
	clone := &StreamManager{
		mappings: make(map[string]Stream),
	}

	for fd, stream := range sm.mappings {
		clone.mappings[fd] = &proxyStream{
			original: stream,
		}
	}
	return clone
}

func NewPipe() (Stream, Stream, error) {
	return os.Pipe()
}
