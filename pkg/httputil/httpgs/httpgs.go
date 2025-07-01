package httpgs

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Set interface {
	Register(name string, server *http.Server, shutdownTimeout time.Duration) Set
	OnMessage(func(name, message string)) Set
	OnError(func(name string, err error)) Set
	ListenAndServe() error
}

type set struct {
	wg          sync.WaitGroup
	m           sync.Mutex
	ctx         context.Context
	onMessage   func(key, message string)
	onError     func(key string, err error)
	items       []*item
	errorsCh    chan error
	done        chan struct{}
	closedByErr chan struct{}
}

type item struct {
	shutdownTimeout time.Duration
	name            string
	server          *http.Server
	done            chan struct{}
}

func NewGracefulShutdown(ctx context.Context) Set {
	return &set{ctx: ctx}
}

func (s *set) OnMessage(messageHandler func(name, message string)) Set {
	s.onMessage = messageHandler
	return s
}

func (s *set) OnError(errorHandler func(name string, err error)) Set {
	s.onError = errorHandler
	return s
}

func (s *set) Register(name string, server *http.Server, shutdownTimeout time.Duration) Set {
	s.m.Lock()
	defer s.m.Unlock()
	item := &item{name: name, server: server, shutdownTimeout: shutdownTimeout, done: make(chan struct{})}
	s.items = append(s.items, item)

	return s
}

func (s *set) ListenAndServe() error {
	s.m.Lock()
	defer s.m.Unlock()

	s.done = make(chan struct{})
	s.closedByErr = make(chan struct{})
	s.errorsCh = make(chan error, len(s.items))

	s.wg.Add(len(s.items))
	for _, item := range s.items {
		go s.listenAndServe(item)
		go s.gracefulShutdown(item)
	}

	go func() {
		fatalCount := 0
		for range s.errorsCh {
			fatalCount++
			if fatalCount == len(s.items) {
				close(s.closedByErr)
			}
		}
	}()

	go func() {
		s.wg.Wait()
		close(s.done)
	}()

	<-s.done
	return nil
}

func (s *set) pushMessage(name, message string) {
	if s.onMessage != nil {
		s.onMessage(name, message)
	}
}

func (s *set) pushError(name string, err error) {
	if s.onError != nil {
		s.onError(name, err)
	}
}

func (s *set) listenAndServe(i *item) {
	s.pushMessage(i.name, fmt.Sprintf("listening on %s", i.server.Addr))

	if err := i.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.pushError(i.name, fmt.Errorf("listen error: %w", err))
		s.errorsCh <- err
		close(i.done)
	}
}

func (s *set) gracefulShutdown(i *item) {
	defer s.wg.Done()
	select {
	case <-s.ctx.Done():
	case <-i.done:
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), i.shutdownTimeout)
	defer cancel()
	s.pushMessage(i.name, fmt.Sprintf("shutting down (%s)...", i.shutdownTimeout))

	if err := i.server.Shutdown(ctx); err != nil {
		s.pushError(i.name, fmt.Errorf("shutting down error: %w", err))
	} else {
		s.pushMessage(i.name, "shutdown (success)")
	}
}
