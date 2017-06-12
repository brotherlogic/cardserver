package main

import (
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
	return true
}

// SaveCardList stores the cardlist
func (s *Server) SaveCardList() {
	s.Save("github.com/brotherlogic/cardserver/cards", s.cards)
}

func main() {
	server := InitServer()
	server.GoServer.KSclient = *keystoreclient.GetClient()
	server.PrepServer()
	server.RegisterServer("cardserver", false)
	server.Serve()
}
