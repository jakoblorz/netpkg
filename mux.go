package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type mux struct {
	ln   net.Listener
	once sync.Once
	wg   sync.WaitGroup

	handlers map[byte]*handle
	lock     sync.Mutex

	rn      byte
	timeout time.Duration
}

func multiplex(ln net.Listener) *mux {
	return &mux{
		ln:       ln,
		handlers: make(map[byte]*handle),
		rn:       '0',
		timeout:  100 * time.Microsecond,
	}
}

func (mux *mux) Close() (err error) {
	mux.once.Do(func() {
		if mux.ln != nil {
			err = mux.ln.Close()
		}
		mux.wg.Wait()
		mux.lock.Lock()
		defer mux.lock.Unlock()
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
	start := time.Now()
	bufConn := &conn{
		conn: c,
		r:    bufio.NewReader(c),
	}
	if err := c.SetReadDeadline(time.Now().Add(mux.timeout)); err != nil {
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
	mux.lock.Lock()
	defer mux.lock.Unlock()
	if h = mux.handlers[hdr]; h == nil && hdr != mux.rn {
		hdr = mux.rn
		h = mux.handlers[hdr]
	}
	if h == nil {
		return fmt.Errorf("unregistered header byte: 0x%02x", hdr)
	}
	log.Printf("Receiving Request %s -> Matching '%s' (0x%02x) (%s) = %s", c.RemoteAddr(), string(hdr), hdr, time.Since(start), h.name)
	h.c <- bufConn
	return nil
}

func (mux *mux) Listen(hdrs []byte, name string) net.Listener {
	h := &handle{
		addr: mux.ln.Addr(),
		name: name,
		c:    make(chan net.Conn),
	}
	mux.lock.Lock()
	defer mux.lock.Unlock()
	for _, hdr := range hdrs {
		mux.handlers[hdr] = h
	}
	return h
}

func (mux *mux) Any(rn byte, name string) net.Listener {
	mux.rn = rn
	return mux.Listen([]byte{rn}, name)
}
