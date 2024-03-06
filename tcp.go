package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
)

const (
	BUFFSIZE int    = 1024
	IPFX     string = "[INFO] "
	EPFX     string = "[ERROR] "
)

var (
	listen bool
	quiet  bool
	addr   string
	errch  chan error = make(chan error, 1)
)

func main() {
	if !parseArgs() {
		flag.PrintDefaults()
		os.Exit(1)
	}

	conn, listener, err := getConn()
	if err != nil {
		log.Fatalln(EPFX+"conexión fallida:", err)
	}
	log.Println(IPFX+"conectado a", conn.RemoteAddr())

	handleConnections(conn, listener)
}

func parseArgs() bool {
	flag.BoolVar(&listen, "l", false, "Se queda a la escucha")
	flag.BoolVar(&quiet, "q", false, "No imprime mensajes de debug")
	flag.StringVar(&addr, "H", ":8080", "Direccion de la forma <IP>:<puerto>")

	flag.Parse()

	if quiet {
		log.SetFlags(0)
		log.SetOutput(io.Discard)
	}

	return flag.Parsed()
}

func getConn() (net.Conn, net.Listener, error) {
	if listen {
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			return nil, nil, err
		}
		log.Println(IPFX+"servidor iniciado en", addr)

		conn, err := listener.Accept()
		return conn, listener, err
	} else {
		conn, err := net.Dial("tcp", addr)
		return conn, nil, err
	}
}

func handleConnections(conn net.Conn, listener net.Listener) {
	buffConn := make([]byte, BUFFSIZE)
	buffOs := make([]byte, BUFFSIZE)
	defer conn.Close()
	if listener != nil {
		listener.Close()
	}

	// conn -> stdout
	go func() {
		for {
			n, err := conn.Read(buffConn)
			if err != nil {
				errch <- err
				return
			}
			os.Stdout.Write(buffConn[:n])
		}
	}()

	// stdin -> conn
	go func() {
		for {
			n, err := os.Stdin.Read(buffOs)
			if err != nil {
				errch <- err
				return
			}
			conn.Write(buffOs[:n])
		}
	}()

	switch err := <-errch; err {
	case io.EOF:
		log.Println(IPFX+"conexión terminada:", err)
	default:
		log.Fatalln(EPFX+"error en la conexión:", err)
	}
}
