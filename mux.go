package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

const (
	master byte = iota
)

type mux struct {
	ln   net.Listener
	once sync.Once
	wg   sync.WaitGroup

	handlers map[byte]*handle

	Timeout time.Duration
}

func multiplex(ln net.Listener) *mux {
	return &mux{
		ln:       ln,
		handlers: make(map[byte]*handle),
		Timeout:  30 * time.Second,
	}
}

func (mux *mux) Close() (err error) {
	mux.once.Do(func() {
		if mux.ln != nil {
			err = mux.ln.Close()
		}
		mux.wg.Wait()
		for _, h := range mux.handlers {
			h.Close()
		}
	})
	return
}

func (mux *mux) Serve() error {
	for {
		conn, err := mux.ln.Accept()
		if err, ok := err.(interface {
			Temporary() bool
		}); ok && err.Temporary() {
			log.Printf("error receiving request: %+v\n", err)
			continue
		}
		if err != nil {
			mux.Close()
			return err
		}
		mux.wg.Add(1)
		go func(conn net.Conn) {
			defer mux.wg.Done()
			if err := mux.handleConn(conn); err != nil {
				conn.Close()
				log.Printf("error handling connection: %+v\n", err)
			}
		}(conn)
	}
}

func (mux *mux) handleConn(c net.Conn) error {
	bufConn := &conn{
		conn: c,
		r:    bufio.NewReader(c),
	}
	if err := c.SetReadDeadline(time.Now().Add(mux.Timeout)); err != nil {
		return fmt.Errorf("set read deadline: %+v", err)
	}
	hdr, err := bufConn.r.ReadByte()
	if err != nil {
		return fmt.Errorf("read header byte: %+v", err)
	} else if err = bufConn.r.UnreadByte(); err != nil {
		return fmt.Errorf("unread header byte: %+v", err)
	}
	if err := c.SetReadDeadline(time.Time{}); err != nil {
		return fmt.Errorf("unset read deadline: %+v", err)
	}
	var h *handle
	if h = mux.handlers[hdr]; h == nil && hdr != master {
		h = mux.handlers[master]
	}
	if h == nil {
		return fmt.Errorf("unregistered header byte: 0x%02x", hdr)
	}
	h.c <- bufConn
	return nil
}

func (mux *mux) Listen(hdrs []byte) net.Listener {
	h := &handle{
		addr: mux.ln.Addr(),
		c:    make(chan net.Conn),
	}
	for _, hdr := range hdrs {
		mux.handlers[hdr] = h
	}
	return h
}

func (mux *mux) Any() net.Listener {
	return mux.Listen([]byte{master})
}
