package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/elmq0022/pub-sub/internal/broker"
	sessioncontroller "github.com/elmq0022/pub-sub/internal/session_controller"
	"github.com/elmq0022/pub-sub/internal/subjectregistry"
)

func main() {
	r := subjectregistry.NewSubjectRegistry()
	b := broker.NewBroker(r, broker.BrokerConfig{
		HeartbeatTickInterval: 30 * time.Second,
		HeartbeatTimeout:      90 * time.Second,
	})
	go b.Run()

	s := sessioncontroller.NewSessionController(b.Input())

	ln, err := net.Listen("tcp", net.JoinHostPort("", "8080"))
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	fmt.Print("listening on :8080")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("accept error")
			continue
		}
		s.Start(conn)
	}
}
