package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
)

var (
	port  int
	host  string
	file  string
	shell string
)

func init() {
	flag.IntVar(&port, "port", 8000, "set the port to listen on")
	flag.StringVar(&host, "host", "0.0.0.0", "set the host to listen on")
	flag.StringVar(&file, "file", "", "set the file to execute")
	flag.StringVar(&shell, "shell", "sh", "set the shell")
}

func listen(network, address string) (*mux, error) {
	ln, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return multiplex(ln), nil
}

func run(shell, file string, r io.Reader, w io.Writer) error {
	cmd := exec.Command(shell, file)
	cmd.Stdin = r
	cmd.Stdout = w
	return cmd.Run()
}

func main() {
	flag.Parse()
	pid, err := os.OpenFile(fmt.Sprintf("%s.pid", file), os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatalf("failed to create pid file: %s", err.Error())
	} else {
		pid.Write([]byte(strconv.Itoa(syscall.Getpid())))
		defer pid.Close()
	}
	sock, err := listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Fatalf("failed to create listener: %s", err.Error())
	}
	defer sock.Close()
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func(shell, file string) {
		defer func() { sig <- syscall.SIGINT }()
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if err := run(shell, file, r.Body, w); err != nil {
				log.Printf("error executing request: %+v\n", err)
			}
		})
		if err := http.Serve(sock.Listen([]byte{'G', 'P', 'D'}), nil); err != nil {
			log.Printf("error listening on port: %+v\n", err)
		}
	}(shell, file)
	go func(shell, file string) {
		defer func() { sig <- syscall.SIGINT }()
		ln := sock.Any()
		for {
			conn, err := ln.Accept()
			if err, ok := err.(interface {
				Temporary() bool
			}); ok && err.Temporary() {
				log.Printf("error recieving request: %+v\n", err)
				continue
			}
			if err != nil {
				log.Printf("error listening on port: %+v\n", err)
				return
			}
			go func(c net.Conn) {
				if err := run(shell, file, c, c); err != nil {
					log.Printf("error executing request: %+v\n", err)
				}
			}(conn)
		}
	}(shell, file)
	go func() {
		defer func() { sig <- syscall.SIGINT }()
		if err := sock.Serve(); err != nil {
			log.Printf("error listening on port: %+v\n", err)
		}
	}()
	for {
		select {
		case s := <-sig:
			log.Printf("signal (%d) received, stopping", s)
			return
		}
	}
}
