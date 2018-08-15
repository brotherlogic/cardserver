package main

import (
	"log"
	"sort"
	"strings"
	"time"

	"github.com/brotherlogic/goserver"

	pb "github.com/brotherlogic/cardserver/card"
	"golang.org/x/net/context"
)

const (
	port = ":50051"
	key  = "/github.com/brotherlogic/cardserver/cards"
)

// Server The server to use
type Server struct {
	*goserver.GoServer
	cards *pb.CardList
}

// InitServer Prepares the server to run
func InitServer() *Server {
	s := &Server{&goserver.GoServer{}, &pb.CardList{}}
	s.Register = s
	return s
}

func (s *Server) removePrefix(prefix string) *pb.CardList {
	filteredCards := &pb.CardList{}
	for _, card := range s.cards.Cards {
		if !strings.HasPrefix(card.Hash, prefix) {
			filteredCards.Cards = append(filteredCards.Cards, card)
		}
	}
	return filteredCards
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
	log.Printf("READING CARDS: %v (%v)", s.cards, s)
	s.removeStaleCards()
	s.cards = s.dedup(s.cards)
	s.sortCards()
	return s.cards, nil
}

// AddCards adds cards to the server
func (s *Server) AddCards(ctx context.Context, in *pb.CardList) (*pb.CardList, error) {
	log.Printf("ADDING CARDS: %v", in)
	s.cards.Cards = append(s.cards.Cards, in.Cards...)
	s.SaveCardList(ctx)
	return s.cards, nil
}

// DeleteCards removes cards from the server
func (s *Server) DeleteCards(ctx context.Context, in *pb.DeleteRequest) (*pb.CardList, error) {
	log.Printf("DELETE: %v", in)
	if in.Hash != "" {
		s.cards = s.remove(in.Hash)
	} else {
		s.cards = s.removePrefix(in.HashPrefix)
	}
	s.SaveCardList(ctx)
	return s.cards, nil
}
