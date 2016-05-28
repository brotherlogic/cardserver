package main

import (
	"log"
	"net"
	"sort"
	"time"

	pb "github.com/brotherlogic/cardserver/card"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const (
	port = ":50051"
)

type server struct {
	cards *pb.CardList
}

func InitServer() server {
	s := server{}
	s.cards = &pb.CardList{}
	return s
}

func (s *server) dedup() {
	log.Printf("Deduping")
}

func (s *server) removeStaleCards() {

	filteredCards := &pb.CardList{}
	for _, card := range s.cards.Cards {
		if card.ExpirationDate <= 0 || card.ExpirationDate >= time.Now().Unix() {
			log.Printf("Not filtering %v (time is %v)", card, time.Now().Unix())
			filteredCards.Cards = append(filteredCards.Cards, card)
		}
	}

	s.cards = filteredCards
}

type ByPriority []*pb.Card

func (a ByPriority) Len() int           { return len(a) }
func (a ByPriority) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByPriority) Less(i, j int) bool { return a[i].Priority < a[j].Priority }

func (s *server) sortCards() {
	sort.Sort(ByPriority(s.cards.Cards))
}

func (s *server) GetCards(ctx context.Context, in *pb.Empty) (*pb.CardList, error) {
	s.removeStaleCards()
	s.sortCards()
	return s.cards, nil
}

func (s *server) AddCards(ctx context.Context, in *pb.CardList) (*pb.CardList, error) {
	s.cards.Cards = append(s.cards.Cards, in.Cards...)
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
