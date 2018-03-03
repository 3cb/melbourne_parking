// Package ssc creates and controls pools of websocket connections --  client and server
package ssc

import (
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Message contains type and payload as a normal websocket message
// ID is added in case user needs to identify source
type Message struct {
	ID      string
	Type    int
	Payload []byte
}

// Pool is a collection of websocket connections combined with
// channels which are used to send and received messages to and from
// the goroutines that control them
type Pool struct {
	mtx        sync.Mutex
	isDraining bool
	rw

	// These fields are set with NewPool parameters
	serverURLs   []string
	pingInterval time.Duration

	// Inbound channel carries messages from caller application to WriteControl() method
	// Messages sent into Inbound channel will be written to ALL connections in stack
	Inbound chan *Message

	// Outbound channel carries messages from ReadControl() method to caller application
	Outbound chan *Message

	// s2p channel carries messages from Read goroutines to ReadControl() method(i.e., socket to pool)
	s2p chan *Message

	// stop channels are used to stop control goroutines
	stopReadControl     chan *sync.WaitGroup
	stopWriteControl    chan *sync.WaitGroup
	stopShutdownControl chan *sync.WaitGroup

	// remove sends *socket to controlShutdown() so that it can send quit signals to that socket's read and write goroutines
	remove chan *socket

	// Shutdown channel carries messages from read/write goroutines to ControlShutdown() goroutine
	shutdown chan *socket

	// allClosed alerts pool.Drain() that all read/write goroutines have been stopped and all websockets have been disconnected
	allClosed chan struct{}
}

// rw contains map of sockets with read and write goroutines
type rw struct {
	mtx   sync.RWMutex
	stack map[*socket]int
}

// NewPool creates a new instance of Pool and returns a pointer to it
func NewPool(urls []string, pingInt time.Duration) *Pool {
	return &Pool{
		isDraining: false,
		rw:         rw{stack: make(map[*socket]int)},

		serverURLs:   urls,
		pingInterval: pingInt,

		Inbound:             make(chan *Message, 200),
		Outbound:            make(chan *Message, 200),
		s2p:                 make(chan *Message, 200),
		stopReadControl:     make(chan *sync.WaitGroup),
		stopWriteControl:    make(chan *sync.WaitGroup),
		stopShutdownControl: make(chan *sync.WaitGroup),
		remove:              make(chan *socket, 200),
		shutdown:            make(chan *socket, 200),
		allClosed:           make(chan struct{}),
	}
}

// Start spins up control goroutines and connects to websockets(if server pool)
// If slice of urls is nil or empty Pool will be created empty and control methods will be launched and waiting
func (p *Pool) Start() error {
	go p.controlRead()
	go p.controlWrite()
	go p.controlShutdown()

	if len(p.serverURLs) > 0 {
		for _, url := range p.serverURLs {
			s := newSocketInstance(url)
			err := s.connectServer(p)
			if err != nil {
				p.Stop()
				return err
			}
			log.Printf("Connected to websocket(%v)\n", url)
		}
	}
	return nil
}

// Count returns the numbers of current websocket connection in stack
func (p *Pool) Count() int {
	p.rw.mtx.RLock()
	count := len(p.rw.stack)
	p.rw.mtx.RUnlock()
	return count
}

// List returns a []string with ids of each socket in stack
func (p *Pool) List() []string {
	list := []string{}
	p.rw.mtx.RLock()
	for s := range p.rw.stack {
		list = append(list, s.id)
	}
	p.rw.mtx.RUnlock()
	return list
}

// Write takes a *Message and writes it to a Socket based on Message.ID
// If Message.ID is an empty string or doesn't match an existing ID in the stack Write will return an error
func (p *Pool) Write(msg *Message) error {
	switch id := msg.ID; id {
	case "":
		return errors.New("cannot send message -- id string is empty")
	default:
		p.rw.mtx.RLock()
		for s := range p.rw.stack {
			if id == s.id {
				s.p2s <- msg
				p.rw.mtx.RUnlock()
				return nil
			}
		}
		p.rw.mtx.RUnlock()
	}
	return errors.New("cannot send message -- socket id does not exist")
}

// WriteAll takes a *Message and writes it to all sockets in stack
func (p *Pool) WriteAll(msg *Message) {
	p.Inbound <- msg
}

// AddClientSocket allows caller to add individual websocket connections to an existing pool
// New connection will adopt existing pool configuration
func (p *Pool) AddClientSocket(id string, upgrader *websocket.Upgrader, w http.ResponseWriter, r *http.Request) error {
	if p.isDraining {
		return errors.New("cannot add new websocket connection -- pool is draining")
	}
	s := newSocketInstance(id)
	err := s.connectClient(p, upgrader, w, r)
	if err != nil {
		return err
	}
	return nil
}

// AddServerSocket allows caller to add individual websocket connections to an existing pool
// New connection will adopt existing pool configuration
func (p *Pool) AddServerSocket(url string) error {
	if p.isDraining {
		return errors.New("cannot add new websocket connection -- pool is draining")
	}
	s := newSocketInstance(url)
	err := s.connectServer(p)
	if err != nil {
		return err
	}
	return nil
}

// RemoveSocket takes an id string and sends *socket to controlShutdown goroutine via remove chan
// returns an error if the id string is empty or is not in the stack
func (p *Pool) RemoveSocket(id string) error {
	switch id {
	case "":
		return errors.New("empty string -- cannot remove socket without proper id")
	default:
		p.rw.mtx.RLock()
		for s := range p.rw.stack {
			if id == s.id {
				s.rQuit <- struct{}{}
				s.wQuit <- struct{}{}
				p.rw.mtx.RUnlock()
				return nil
			}
		}
		p.rw.mtx.RUnlock()
	}
	return errors.New("id not found in stack -- cannot remove")
}

// Stop shuts down all read and write goroutines as well as all control goroutines
func (p *Pool) Stop() {
	p.mtx.Lock()
	p.isDraining = true
	p.mtx.Unlock()

	p.rw.mtx.RLock()
	if len(p.rw.stack) > 0 {
		for s := range p.rw.stack {
			s.rQuit <- struct{}{}
			s.wQuit <- struct{}{}
		}
	}
	p.rw.mtx.RUnlock()

	<-p.allClosed

	wg := &sync.WaitGroup{}
	wg.Add(3)
	p.stopReadControl <- wg
	p.stopWriteControl <- wg
	p.stopShutdownControl <- wg
	wg.Wait()
}
