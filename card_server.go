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

type server struct {
     cards *pb.CardList
}

func InitServer() (server) {
     s := server{}
     s.cards = &pb.CardList{}

     return s
}

func (s *server) GetCards (ctx context.Context, in *pb.Empty) (*pb.CardList, error) {
     return s.cards, nil
}

func (s *server) AddCards (ctx context.Context, in *pb.CardList) (*pb.CardList, error) {
     s.cards.Cards = append(s.cards.Cards, in.Cards...)
     log.Printf("Received %v", s.cards.Cards)
     return s.cards, nil
}

func main() {
     lis, err := net.Listen("tcp", port)
     if err != nil {
     	log.Fatalf("Failed to listen on %v", err)
     }

     s := grpc.NewServer()
     server := InitServer()
     pb.RegisterCardServiceServer(s, &server)
     s.Serve(lis)
}