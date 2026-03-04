package main

import (
	"fmt"
	"log"
	"net"

	"github.com/elmq0022/pub-sub/internal/broker"
	"github.com/elmq0022/pub-sub/internal/config"
	"github.com/elmq0022/pub-sub/internal/sessioncontroller"
	"github.com/elmq0022/pub-sub/internal/subjectregistry"
)

func main() {
	r := subjectregistry.NewSubjectRegistry()
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatal(err)
	}

	b := broker.NewBroker(r, cfg)
	go b.Run()

	s := sessioncontroller.NewSessionController(b.Input())

	ln, err := net.Listen("tcp", net.JoinHostPort("", cfg.Port))
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	fmt.Printf("listening on :%s", cfg.Port)

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("accept error")
			continue
		}
		s.Start(conn)
	}
}
