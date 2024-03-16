package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
)

const (
	IPFX string = "[INFO] "
	EPFX string = "[ERROR] "
)

var (
	// argumentos de línea de órdenes
	listen   bool
	quiet    bool
	addr     string
	port     uint
	buffsize uint
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

	switch err = handleConnections(conn, listener); err {
	case io.EOF:
		log.Println(IPFX+"conexión terminada:", err)
	default:
		log.Fatalln(EPFX+"error en la conexión:", err)
	}
}

func parseArgs() bool {
	flag.BoolVar(&listen, "l", false, "Se queda a la escucha")
	flag.BoolVar(&quiet, "q", false, "No imprime mensajes de debug")
	flag.StringVar(&addr, "H", "", "Direccion IP")
	flag.UintVar(&buffsize, "b", 1, "Tamaño del búffer (en KB)")
	flag.UintVar(&port, "p", 8080, "Puerto")

	flag.Parse()

	if quiet {
		log.SetFlags(0)
		log.SetOutput(io.Discard)
	}

	return flag.Parsed()
}

func getConn() (net.Conn, net.Listener, error) {
	address := fmt.Sprintf("%s:%v", addr, port)

	if listen {
		listener, err := net.Listen("tcp", address)
		if err != nil {
			return nil, nil, err
		}
		log.Println(IPFX+"servidor iniciado en", address)

		conn, err := listener.Accept()
		return conn, listener, err
	} else {
		conn, err := net.Dial("tcp", address)
		return conn, nil, err
	}
}

func handleConnections(conn net.Conn, listener net.Listener) error {
	var errch = make(chan error, 1)
	buffConn := make([]byte, 1024*buffsize)
	buffOs := make([]byte, 1024*buffsize)
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

	return <-errch
}
