package ochanw

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
		for {
			select {
			case w, ok := <-o.in:
				if !ok {
					close(o.done)
					return
				}

				w.mu.Lock()
				if w.state == background {
					// ロック不要
					w.state = toCurrent
				}
				w.mu.Unlock()

				func() {
					for {
						select {
						case <-w.done:
							if len(w.buffer) > 0 {
								w.parent.out.Write(w.buffer)
							}
							return
						}
					}
				}()
			}
		}
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
func (o *Ow) GetW() *WriteCloser {
	next := &WriteCloser{
		parent: o,
		state:  background,
		done:   make(chan struct{}, 1),
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

// Write ...
func (w *WriteCloser) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	switch w.state {
	case background:
		// background 時はため込む
		w.buffer = append(w.buffer, p...)
		return len(p), nil
	case toCurrent:
		// lock しつつため込んだものを出力し、切り替える
		_, err := w.parent.out.Write(w.buffer)
		if err != nil {
			return 0, err
		}
		w.buffer = w.buffer[0:0]
		w.state = current
	case current:
		// 直接 write する
	}

	return w.parent.out.Write(p)
}

// Close ...
func (w *WriteCloser) Close() error {
	w.done <- struct{}{}
	return nil
}
