package ow

import (
	"io"
	"sync"
)

// Ow is a structure for controlling the output order of io.Writers
type Ow struct {
	out  io.Writer
	in   chan *WriteCloser
	done chan struct{}
}

type owState int

const (
	background owState = iota + 1
	toCurrent
	current
)

// Option ...
type Option func(*options)

type options struct {
	size int
}

var defaultOption = options{
	size: 100,
}

// WithSize ...
func WithSize(s int) Option {
	return func(o *options) {
		o.size = s
	}
}

// New returns a new Ow struct
func New(out io.Writer, opt ...Option) *Ow {
	opts := defaultOption
	for _, f := range opt {
		f(&opts)
	}

	o := &Ow{
		out:  out,
		in:   make(chan *WriteCloser, opts.size),
		done: make(chan struct{}),
	}

	go func(o *Ow) {
		for w := range o.in {

			w.mu.Lock()
			if w.state == background {
				w.state = toCurrent
			}
			w.mu.Unlock()

			for range w.done {
				if len(w.buffer) > 0 {
					w.parent.out.Write(w.buffer)
				}
			}
		}
		close(o.done)
	}(o)

	return o
}

// Wait blocks until it retrieves data from all Ow.
func (o *Ow) Wait() error {
	close(o.in)
	<-o.done
	return nil
}

// GetW returns a next io.WriteCloser.
func (o *Ow) GetW(opt ...Option) *WriteCloser {
	opts := defaultOption
	for _, f := range opt {
		f(&opts)
	}

	next := &WriteCloser{
		buffer: make([]byte, 0, opts.size),
		state:  background,
		parent: o,
		done:   make(chan struct{}, 1),
		mu:     sync.Mutex{},
	}
	o.in <- next
	return next
}

// WriteCloser ...
type WriteCloser struct {
	buffer []byte
	state  owState
	parent *Ow
	done   chan struct{}
	mu     sync.Mutex
}

// Write writes to ow's io.Writer.
func (w *WriteCloser) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	switch w.state {
	case background:
		w.buffer = append(w.buffer, p...)
		return len(p), nil
	case toCurrent:
		_, err := w.parent.out.Write(w.buffer)
		if err != nil {
			return 0, err
		}
		w.buffer = w.buffer[0:0]
		w.state = current
	case current:
	}

	return w.parent.out.Write(p)
}

// Close closes ow.WriteCloser.
func (w *WriteCloser) Close() error {
	w.done <- struct{}{}
	close(w.done)
	return nil
}
