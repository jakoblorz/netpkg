package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var (
	port  int
	host  string
	name  string
	token string
	chars = []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
)

func init() {
	flag.IntVar(&port, "p", 8000, "specify the port")
	flag.StringVar(&host, "h", "0.0.0.0", "specify the host")
	flag.StringVar(&name, "c", "sh", "name of the program to execute")
	flag.StringVar(&token, "t", "", "secure the api with a token; set 'n' if no token is required")
}

func listen(network, address string) (*mux, error) {
	ln, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return multiplex(ln), nil
}

func run(name string, args []string, r io.Reader, w io.Writer) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = r
	cmd.Stdout = w
	return cmd.Run()
}

func main() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
	if token == "" {
		var builder strings.Builder
		for i := 0; i < 32; i++ {
			builder.WriteRune(chars[rand.Intn(len(chars))])
		}
		token = fmt.Sprintf("0x%s", builder.String())
		log.Printf("Using Generated Token %s\n", token)
	} else if token == "n" {
		token = ""
		log.Printf("Selected Not To Use Any Tokens\n")
	} else {
		log.Printf("Using Configured Token %s\n", token)
	}
	args := os.Args[len(os.Args)-flag.NArg():]
	argsMsg := fmt.Sprintf("Using Configured Command \"%s", name)
	for _, arg := range args {
		argsMsg = fmt.Sprintf("%s %s", argsMsg, arg)
	}
	log.Printf("%s\"\n", argsMsg)
	pid, err := os.OpenFile("netfn.pid", os.O_RDWR|os.O_CREATE, 0666)
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
	go func(name string, args []string) {
		defer func() { sig <- syscall.SIGINT }()
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if err := run(name, args, r.Body, w); err != nil {
				log.Printf("error executing request: %+v\n", err)
			}
		})
		if err := http.Serve(sock.Listen([]byte{'G', 'P', 'D'}, "net/http"), nil); err != nil {
			log.Printf("error listening on port: %+v\n", err)
		}
	}(name, args)
	go func(shell string, args []string) {
		defer func() { sig <- syscall.SIGINT }()
		var rn byte = '0'
		if token != "" {
			rn = token[0]
		}
		ln := sock.Any(rn, "net/tcp")
		defer ln.Close()
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
				if err := run(name, args, c, c); err != nil {
					log.Printf("error executing request: %+v\n", err)
					return
				}
				c.Close()
			}(conn)
		}
	}(name, args)
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
