package main

import (
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

func main() {
	server := InitServer()
	server.PrepServer()
	server.RegisterServer("cardserver", false)
	server.Serve()
}
