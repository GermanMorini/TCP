package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

const (
	IPFX string = "[INFO] "
	EPFX string = "[ERROR] "
)

var (
	errch chan error = make(chan error, 1)

	// argumentos de línea de comandos
	listen   bool
	quiet    bool
	addr     string
	proto    string
	port     uint
	buffsize uint
)

// implementación de net.Listener para una conexión UDP
type UDPListener struct{}

func (l *UDPListener) Accept() (net.Conn, error) {
	conn, err := net.ListenPacket("udp", addr)
	return &UDPConn{conn: conn}, err
}
func (l *UDPListener) Close() error   { return nil }
func (l *UDPListener) Addr() net.Addr { return nil }

// implementación de net.Conn para una conexión UDP
type UDPConn struct {
	conn net.PacketConn
	addr net.Addr
}

func (c *UDPConn) Read(p []byte) (int, error) {
	var n int
	var err error

	n, c.addr, err = c.conn.ReadFrom(p)
	return n, err
}

func (c *UDPConn) Write(p []byte) (int, error) {
	n, err := c.conn.WriteTo(p, c.addr)
	return n, err
}

func (c *UDPConn) Close() error                       { return c.conn.Close() }
func (c *UDPConn) LocalAddr() net.Addr                { return c.conn.LocalAddr() }
func (c *UDPConn) RemoteAddr() net.Addr               { return c.conn.LocalAddr() }
func (c *UDPConn) SetDeadline(t time.Time) error      { return nil }
func (c *UDPConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *UDPConn) SetWriteDeadline(t time.Time) error { return nil }

func main() {
	if !parseArgs() {
		flag.PrintDefaults()
		os.Exit(1)
	}

	var err error
	var listener net.Listener
	var conn net.Conn

	if listen {
		listener, err = getListener()
		if err != nil {
			log.Fatalln(EPFX+"conexión fallida:", err)
		}
		log.Println(IPFX+"servidor iniciado en", addr)
		defer listener.Close()
	}

	conn, err = getConn(listener)
	if err != nil {
		log.Fatalln(EPFX+"conexión fallida:", err)
	}
	log.Printf("%sconectado a %s (%s)\n",
		IPFX,
		conn.RemoteAddr().String(),
		strings.Split(conn.LocalAddr().String(), ":")[1],
	)
	defer conn.Close()

	go readWriteLoop(conn, os.Stdout)
	go readWriteLoop(os.Stdin, conn)

	switch err = <-errch; err {
	case io.EOF:
		log.Println(IPFX+"conexión terminada:", err)
	default:
		log.Fatalln(EPFX+"error en la conexión:", err)
	}
}

func parseArgs() bool {
	var udp bool

	flag.BoolVar(&listen, "l", false, "Se queda a la escucha")
	flag.BoolVar(&quiet, "q", false, "No imprime mensajes de debug")
	flag.BoolVar(&udp, "u", false, "Utiliza UDP en lugar de TCP")
	flag.StringVar(&addr, "H", "", "Direccion IP")
	flag.UintVar(&buffsize, "b", 1, "Tamaño del búffer (en KB)")
	flag.UintVar(&port, "p", 8080, "Puerto")
	flag.Parse()

	if udp {
		proto = "udp"
	} else {
		proto = "tcp"
	}

	if quiet {
		log.SetFlags(0)
		log.SetOutput(io.Discard)
	} else {
		log.SetOutput(os.Stderr)
	}

	addr = fmt.Sprintf("%s:%v", addr, port)

	return flag.Parsed()
}

func getListener() (net.Listener, error) {
	if proto == "udp" {
		return &UDPListener{}, nil
	} else {
		return net.Listen("tcp", addr)
	}
}

func getConn(ln net.Listener) (net.Conn, error) {
	if ln != nil {
		return ln.Accept()
	} else {
		return net.Dial(proto, addr)
	}
}

func readWriteLoop(in io.Reader, out io.Writer) {
	buffer := make([]byte, 1024*buffsize)
	for {
		n, err := in.Read(buffer)
		if err != nil {
			errch <- err
			return
		}
		_, err = out.Write(buffer[:n])
		if err != nil {
			errch <- err
			return
		}
	}
}
