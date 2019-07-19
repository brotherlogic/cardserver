package main

import (
	"flag"
	"io/ioutil"
	"log"

	"github.com/brotherlogic/keystore/client"
	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/brotherlogic/cardserver/card"
	pbgs "github.com/brotherlogic/goserver/proto"
)

// DoRegister Registers this server
func (s *Server) DoRegister(server *grpc.Server) {
	pb.RegisterCardServiceServer(server, s)
}

// ReportHealth Determines if the server is healthy
func (s *Server) ReportHealth() bool {
	return true
}

// Shutdown the server
func (s *Server) Shutdown(ctx context.Context) error {
	return nil
}

// Mote promotes this server
func (s *Server) Mote(ctx context.Context, master bool) error {
	if master {
		err := s.prepareList(ctx)
		return err
	}
	return nil
}

// GetState gets the state of server
func (s Server) GetState() []*pbgs.State {
	return []*pbgs.State{}
}

// SaveCardList stores the cardlist
func (s *Server) SaveCardList(ctx context.Context) {
	log.Printf("STARTED SAVE")
	s.Save(ctx, key, s.cards)
	log.Printf("FINISHED SAVE")
}

func (s *Server) prepareList(ctx context.Context) error {
	cl := &pb.CardList{}
	rc, _, err := s.Read(ctx, key, cl)
	log.Printf("READ %v", rc)
	if err != nil {
		log.Printf("Failed to read cards! %v", err)
		return err
	}

	s.cards = rc.(*pb.CardList)
	log.Printf("SERVING: %v (%v)", s.cards, s)
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
	server.GoServer.KSclient = *keystoreclient.GetClient(server.DialMaster)
	server.PrepServer()
	server.RegisterServer("cardserver", false)

	server.Serve()
}
