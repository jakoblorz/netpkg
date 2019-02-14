package main

import (
	"errors"
	"net"
	"sync"
)

type handle struct {
	addr net.Addr
	name string
	c    chan net.Conn
	once sync.Once
}

func (h *handle) Accept() (c net.Conn, err error) {
	conn, ok := <-h.c
	if !ok {
		return nil, errors.New("network connection closed")
	}
	return conn, nil
}

func (h *handle) Close() error {
	h.once.Do(func() { close(h.c) })
	return nil
}

func (h *handle) Addr() net.Addr {
	return h.addr
}
