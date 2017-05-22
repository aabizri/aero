package fmtp_test

import (
	"flag"
	"log"
	"os"

	"net"

	"github.com/aabizri/aero/fmtp"
)

var (
	defaultClient *fmtp.Client

	address string
	mock    bool
)

func init() {
	flag.StringVar(&address, "addr", "127.0.0.1:9050", "Address to test against")
	flag.BoolVar(&mock, "mock", true, "Mock the remote endpoint")
	flag.Parse()

	c, err := fmtp.NewClient("localID")
	if err != nil {
		log.Fatalf("error while creating new client: %s", err)
	}
	defaultClient = c

	mockListener(address)
}

func mockListener(addr string) {
	log := log.New(os.Stdout, "mock remote> ", 0)

	laddr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.Fatalf("error resolving TCP address: %v", err)
	}
	listener, err := net.ListenTCP("tcp", laddr)
	if err != nil {
		log.Fatalf("error creating listener: %v", err)
	}
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Fatalf("error accepting connection: %v", err)
			}
			log.Println("new connection formed...")
			go func() {
				for {
					b := make([]byte, 80)
					conn.Read(b)
					log.Printf("received: %s", b)
				}
			}()
		}
	}()
}
