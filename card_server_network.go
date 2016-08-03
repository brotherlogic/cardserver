package main

import (
	"google.golang.org/grpc"

	pb "github.com/brotherlogic/cardserver/card"
)

// DoRegister Registers this server
func (s *Server) DoRegister(server *grpc.Server) {
	pb.RegisterCardServiceServer(server, s)
}

func main() {
	server := InitServer()
	server.PrepServer()
	server.RegisterServer("cardserver", false)
	server.Serve()
}
