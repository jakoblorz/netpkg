package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

var (
	url   string
	port  int
	host  string
	file  string
	shell string
)

func init() {
	flag.StringVar(&url, "url", "/", "set a url")
	flag.IntVar(&port, "port", 8000, "set the port to listen on")
	flag.StringVar(&host, "host", "0.0.0.0", "set the host to listen on")
	flag.StringVar(&file, "file", "", "set the file to execute")
	flag.StringVar(&shell, "shell", "sh", "set the shell")
}

func main() {
	flag.Parse()
	http.HandleFunc(url, func(w http.ResponseWriter, r *http.Request) {
		cmd := exec.Command(shell, file)
		cmd.Stdin = r.Body
		cmd.Stdout = w
		err := cmd.Run()
		if err != nil {
			log.Print(err)
		}
	})
	go func() {
		if err := http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), nil); err != nil {
			log.Printf("error listening on http port: %+v\n", err)
		}
	}()
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case s := <-sig:
			log.Printf("signal (%d) received, stopping", s)
			return
		}
	}
}
