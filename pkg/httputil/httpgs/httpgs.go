package httpgs

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// Set is an interface for managing and gracefully shutting down multiple HTTP servers.
type Set interface {
	// Register adds a server and its shutdown timeout to the Set.
	Register(name string, server *http.Server, shutdownTimeout time.Duration) Set
	// OnMessage sets a callback to handle informational messages.
	OnMessage(func(name, message string)) Set
	// OnError sets a callback to handle errors from the servers.
	OnError(func(name string, err error)) Set
	// ListenAndServe starts all registered servers and manages their graceful shutdown.
	ListenAndServe() error
}

// set is a type used to manage multiple HTTP servers with graceful shutdown capabilities.
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

// item represents a server instance with configurable shutdown behavior and a communication channel for state updates.
type item struct {
	shutdownTimeout time.Duration
	name            string
	server          *http.Server
	done            chan struct{}
}

// NewGracefulShutdown creates a `Set` instance to manage graceful server shutdowns using the provided context.
func NewGracefulShutdown(ctx context.Context) Set {
	return &set{ctx: ctx}
}

// OnMessage sets a callback function to handle messages from the set.
func (s *set) OnMessage(messageHandler func(name, message string)) Set {
	s.onMessage = messageHandler
	return s
}

// OnError sets a callback to be invoked when an error occurs.
func (s *set) OnError(errorHandler func(name string, err error)) Set {
	s.onError = errorHandler
	return s
}

// Register adds a server to the set.
func (s *set) Register(name string, server *http.Server, shutdownTimeout time.Duration) Set {
	s.m.Lock()
	defer s.m.Unlock()
	item := &item{name: name, server: server, shutdownTimeout: shutdownTimeout, done: make(chan struct{})}
	s.items = append(s.items, item)

	return s
}

// ListenAndServe starts all registered servers, manages their lifecycle, and blocks until servers complete or encounter errors.
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

// pushMessage invokes the onMessage callback if defined, passing the provided server name and message as parameters.
func (s *set) pushMessage(name, message string) {
	if s.onMessage != nil {
		s.onMessage(name, message)
	}
}

// pushError invokes the onError callback if defined, passing the provided server name and error as arguments.
func (s *set) pushError(name string, err error) {
	if s.onError != nil {
		s.onError(name, err)
	}
}

// listenAndServe starts the HTTP server for the given item and handles errors, notifying via callbacks or error channels.
func (s *set) listenAndServe(i *item) {
	s.pushMessage(i.name, fmt.Sprintf("listening on %s", i.server.Addr))

	if err := i.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		s.pushError(i.name, fmt.Errorf("listen error: %w", err))
		s.errorsCh <- err
		close(i.done)
	}
}

// gracefulShutdown gracefully stops the provided server instance within the defined shutdown timeout or on context cancellation.
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
