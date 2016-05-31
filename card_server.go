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

// Server The server to use
type Server struct {
	cards *pb.CardList
}

// InitServer Prepares the server to run
func InitServer() Server {
	s := Server{}
	s.cards = &pb.CardList{}
	return s
}

func (s *Server) remove(hash string) *pb.CardList {
	filteredCards := &pb.CardList{}
	for _, card := range s.cards.Cards {
		if card.Hash != hash {
			filteredCards.Cards = append(filteredCards.Cards, card)
		}
	}
	return filteredCards
}

func (s *Server) dedup(list *pb.CardList) *pb.CardList {
	filteredCards := &pb.CardList{}
	var seen map[string]bool
	seen = make(map[string]bool)
	for _, card := range list.Cards {
		if _, ok := seen[card.Hash]; !ok {
			filteredCards.Cards = append(filteredCards.Cards, card)
			seen[card.Hash] = true
		}
	}
	return filteredCards
}

func (s *Server) removeStaleCards() {

	filteredCards := &pb.CardList{}
	for _, card := range s.cards.Cards {
		if card.ExpirationDate <= 0 || card.ExpirationDate >= time.Now().Unix() {
			log.Printf("Not filtering %v (time is %v)", card, time.Now().Unix())
			filteredCards.Cards = append(filteredCards.Cards, card)
		}
	}

	s.cards = filteredCards
}

type byPriority []*pb.Card

func (a byPriority) Len() int           { return len(a) }
func (a byPriority) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a byPriority) Less(i, j int) bool { return a[i].Priority > a[j].Priority }

func (s *Server) sortCards() {
	sort.Sort(byPriority(s.cards.Cards))
}

// GetCards gets the card list on the server
func (s *Server) GetCards(ctx context.Context, in *pb.Empty) (*pb.CardList, error) {
	s.removeStaleCards()
	s.cards = s.dedup(s.cards)
	s.sortCards()
	return s.cards, nil
}

// AddCards adds cards to the server
func (s *Server) AddCards(ctx context.Context, in *pb.CardList) (*pb.CardList, error) {
	s.cards.Cards = append(s.cards.Cards, in.Cards...)
	return s.cards, nil
}

// DeleteCards removes cards from the server
func (s *Server) DeleteCards(ctx context.Context, in *pb.DeleteRequest) (*pb.CardList, error) {
	s.cards = s.remove(in.Hash)
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
