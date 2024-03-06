package main

import (
	"flag"
	"io"
	"log"
	"net"
	"os"
)

const (
	iPfx string = "[INFO] "
	ePfx string = "[ERROR] "
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

	conn, err := getConn()
	if err != nil {
		log.Println(ePfx+"conexión fallida:", err)
		os.Exit(2)
	}
	log.Println(iPfx+"conectado a", conn.RemoteAddr())

	go handle(conn, os.Stdin)
	go handle(os.Stdout, conn)

	switch err = <-errch; err {
	case io.EOF:
		log.Println(iPfx+"conexión terminada:", err)
	default:
		log.Fatalln(ePfx+"error en la conexión:", err)
	}
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

func getConn() (net.Conn, error) {
	if listen {
		listener, err := net.Listen("tcp", addr)
		if err != nil {
			return nil, err
		}
		log.Println(iPfx+"servidor iniciado en", addr)

		conn, err := listener.Accept()
		return conn, err
	} else {
		conn, err := net.Dial("tcp", addr)
		return conn, err
	}
}

func handle(out io.WriteCloser, in io.ReadCloser) {
	buff := make([]byte, 1024)
	defer out.Close()
	defer in.Close()

	for {
		n, err := in.Read(buff)
		if err != nil {
			errch <- err
			return
		}

		out.Write(buff[:n])
	}
}
