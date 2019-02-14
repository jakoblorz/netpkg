package main

import (
	"bufio"
	"net"
	"time"
)

type conn struct {
	conn net.Conn
	r    *bufio.Reader
}

func (c *conn) Read(b []byte) (n int, err error) {
	return c.r.Read(b)
}

func (c *conn) Write(b []byte) (n int, err error) {
	return c.conn.Write(b)
}

func (c *conn) Close() error {
	return c.conn.Close()
}

func (c *conn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *conn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *conn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *conn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *conn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}
