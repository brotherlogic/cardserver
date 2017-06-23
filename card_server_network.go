package main

import (
	"context"
	"log"

	"github.com/brotherlogic/keystore/client"
	"google.golang.org/grpc"

	pb "github.com/brotherlogic/cardserver/card"
	pbdi "github.com/brotherlogic/discovery/proto"
)

// DoRegister Registers this server
func (s *Server) DoRegister(server *grpc.Server) {
	pb.RegisterCardServiceServer(server, s)
}

// ReportHealth Determines if the server is healthy
func (s *Server) ReportHealth() bool {
	return true
}

// Mote promotes this server
func (s *Server) Mote(master bool) error {
	if master {
		s.prepareList()
	}
	return nil
}

// SaveCardList stores the cardlist
func (s *Server) SaveCardList() {
	s.Save(key, s.cards)
}

func findServer(name string) (string, int) {
	conn, err := grpc.Dial("192.168.86.64:50055", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Cannot reach discover server: %v (trying to discover %v)", err, name)
	}
	defer conn.Close()

	registry := pbdi.NewDiscoveryServiceClient(conn)
	rs, err := registry.ListAllServices(context.Background(), &pbdi.Empty{})

	if err != nil {
		log.Fatalf("Failure to list: %v", err)
	}

	for _, r := range rs.Services {
		if r.Name == name {
			log.Printf("%v -> %v", name, r)
			return r.Ip, int(r.Port)
		}
	}

	log.Printf("No Cardserver running")

	return "", -1
}

func (s *Server) prepareList() {
	cl := &pb.CardList{}
	rc, err := s.Read(key, cl)
	log.Printf("READ %v", rc)
	if err != nil {
		log.Printf("Failed to read cards! %v", err)
	} else {
		s.cards = rc.(*pb.CardList)
	}
	log.Printf("SERVING: %v (%v)", s.cards, s)
}

func main() {
	server := InitServer()
	server.GoServer.KSclient = *keystoreclient.GetClient(findServer)
	server.PrepServer()
	server.RegisterServer("cardserver", false)
	server.prepareList()
	log.Printf("SERVING WITH %v (%v)", server.cards, server)
	server.Serve()
}
