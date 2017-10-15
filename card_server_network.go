package main

import (
	"context"
	"flag"
	"io/ioutil"
	"log"
	"time"

	"github.com/brotherlogic/keystore/client"
	"google.golang.org/grpc"

	pb "github.com/brotherlogic/cardserver/card"
)

// DoRegister Registers this server
func (s *Server) DoRegister(server *grpc.Server) {
	pb.RegisterCardServiceServer(server, s)
}

// ReportHealth Determines if the server is healthy
func (s *Server) ReportHealth() bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	_, err := s.GetCards(ctx, &pb.Empty{})
	return err == nil
}

// Mote promotes this server
func (s *Server) Mote(master bool) error {
	if master {
		err := s.prepareList()
		return err
	}
	return nil
}

// SaveCardList stores the cardlist
func (s *Server) SaveCardList() {
	log.Printf("STARTED SAVE")
	s.Save(key, s.cards)
	log.Printf("FINISHED SAVE")
}

func (s *Server) prepareList() error {
	t := time.Now()
	cl := &pb.CardList{}
	rc, err := s.Read(key, cl)
	log.Printf("READ %v", rc)
	if err != nil {
		log.Printf("Failed to read cards! %v", err)
		return err
	}

	s.cards = rc.(*pb.CardList)
	log.Printf("SERVING: %v (%v)", s.cards, s)
	s.LogFunction("prepareList", t)
	return nil
}

func main() {
	var quiet = flag.Bool("quiet", true, "Show all output")
	flag.Parse()
	//Turn off logging
	if *quiet {
		log.SetFlags(0)
		log.SetOutput(ioutil.Discard)
	}
	log.Printf("Logging is on!")

	server := InitServer()
	server.GoServer.KSclient = *keystoreclient.GetClient(server.GetIP)
	server.PrepServer()
	server.RegisterServer("cardserver", false)
	if server.prepareList() != nil {
		log.Printf("Unable to find cardserver details")
		return
	}
	log.Printf("SERVING WITH %v (%v)", server.cards, server)
	server.Serve()
}
