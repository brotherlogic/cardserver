package main

import (
       "log"
       "net"
       
       "golang.org/x/net/context"
       "google.golang.org/grpc"
       pb "github.com/brotherlogic/cardserver/card"
)

const (
      port = ":50051"
)

type server struct {}

func (s *server) GetCards (ctx context.Context, in *pb.Empty) (*pb.CardList, error) {
     return &pb.CardList{}, nil
}

func main() {
     lis, err := net.Listen("tcp", port)
     if err != nil {
     	log.Fatalf("Failed to listen on %v", err)
     }

     s := grpc.NewServer()
     pb.RegisterCardServiceServer(s, &server{})
     s.Serve(lis)
}