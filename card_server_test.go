package main

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/brotherlogic/keystore/client"

	pb "github.com/brotherlogic/cardserver/card"
	"golang.org/x/net/context"
)

func InitTestServer(clear bool) *Server {
	if clear {
		os.RemoveAll(".testing")
	}

	s := InitServer()
	s.GoServer.KSclient = *keystoreclient.GetTestClient(".testing/")
	s.prepareList()
	return s
}

func TestPriority(t *testing.T) {
	card1 := pb.Card{
		Priority: 10,
		Hash:     "10",
	}
	card2 := pb.Card{
		Priority: 50,
		Hash:     "50",
	}
	card3 := pb.Card{
		Priority: 5,
		Hash:     "5",
	}

	cardlist := pb.CardList{}
	cardlist.Cards = append(cardlist.Cards, &card1)
	cardlist.Cards = append(cardlist.Cards, &card2)
	cardlist.Cards = append(cardlist.Cards, &card3)

	s := InitTestServer(true)
	cards, err := s.AddCards(context.Background(), &cardlist)
	if err != nil {
		t.Errorf("Error adding cards %v", err)
	}
	cards, err = s.GetCards(context.Background(), &pb.Empty{})
	if err != nil {
		t.Errorf("Error getting cards %v", err)
	}

	fmt.Printf("CARDS = %v", cards.Cards)
	for i := 1; i < len(cards.Cards); i++ {
		if cards.Cards[i].Priority > cards.Cards[i-1].Priority {
			t.Errorf("Cards are not priority sorted %v -> %v", cards.Cards[i], cards.Cards[i-1])
		}
	}
}

func TestDedup(t *testing.T) {
	card1 := pb.Card{
		Hash: "madeup",
	}
	card2 := pb.Card{
		Hash: "ditto",
	}
	card3 := pb.Card{
		Hash: "madeup",
	}

	cardlist := pb.CardList{}
	cardlist.Cards = append(cardlist.Cards, &card1)
	cardlist.Cards = append(cardlist.Cards, &card2)
	cardlist.Cards = append(cardlist.Cards, &card3)

	s := InitTestServer(true)
	cards, err := s.AddCards(context.Background(), &cardlist)
	if err != nil {
		t.Errorf("Error adding cards %v", err)
	}
	cards, err = s.GetCards(context.Background(), &pb.Empty{})
	if err != nil {
		t.Errorf("Error getting cards %v", err)
	}
	if len(cards.Cards) != 2 {
		t.Errorf("Cards have not been deduped")
	}
}

func TestAdd(t *testing.T) {
	card := pb.Card{}
	s := InitTestServer(true)

	cardlist := pb.CardList{}
	cardlist.Cards = append(cardlist.Cards, &card)

	cards, err := s.AddCards(context.Background(), &cardlist)

	if err != nil {
		t.Errorf("Error in adding a card: %v", err)
	}

	if len(cards.Cards) != 1 {
		t.Errorf("Not enough cards: %v", len(cards.Cards))
	}
}

func TestDeletePrefix(t *testing.T) {
	card := pb.Card{}
	card.Hash = "deleteone"
	card2 := pb.Card{}
	card2.Hash = "deletetwo"
	card3 := pb.Card{}
	card3.Hash = "savethis"

	s := InitTestServer(true)

	cardlist := pb.CardList{}
	cardlist.Cards = append(cardlist.Cards, &card)
	cardlist.Cards = append(cardlist.Cards, &card2)
	cardlist.Cards = append(cardlist.Cards, &card3)

	cards, err := s.AddCards(context.Background(), &cardlist)
	if err != nil {
		t.Errorf("Error in adding cards: %v", err)
	}

	if len(cards.Cards) != 3 {
		t.Errorf("Error adding card: %v", cards)
	}

	deleteReq := pb.DeleteRequest{}
	deleteReq.HashPrefix = "delete"
	cards, err = s.DeleteCards(context.Background(), &deleteReq)
	if err != nil {
		t.Errorf("Fail to delete cards %v", err)
	}
	cards, err = s.GetCards(context.Background(), &pb.Empty{})
	if err != nil {
		t.Errorf("Fail to get cards %v", err)
	}

	if len(cards.Cards) != 1 {
		t.Errorf("Card has not been deleted correctly: %v:%v", len(cards.Cards), cards.Cards)
	}

	if cards.Cards[0].Hash != "savethis" {
		t.Errorf("Card has not been retained correctly: %v", cards.Cards)
	}
}

func TestDeleteAll(t *testing.T) {
	card := pb.Card{}
	card.Hash = "deleteone"
	card2 := pb.Card{}
	card2.Hash = "deletetwo"
	card3 := pb.Card{}
	card3.Hash = "savethis"

	s := InitTestServer(true)

	cardlist := pb.CardList{}
	cardlist.Cards = append(cardlist.Cards, &card)
	cardlist.Cards = append(cardlist.Cards, &card2)
	cardlist.Cards = append(cardlist.Cards, &card3)

	cards, err := s.AddCards(context.Background(), &cardlist)
	if err != nil {
		t.Errorf("Error in adding cards: %v", err)
	}

	if len(cards.Cards) != 3 {
		t.Errorf("Error adding card: %v", cards)
	}

	deleteReq := pb.DeleteRequest{}
	deleteReq.HashPrefix = ""
	cards, err = s.DeleteCards(context.Background(), &deleteReq)
	if err != nil {
		t.Errorf("Fail to delete cards %v", err)
	}
	cards, err = s.GetCards(context.Background(), &pb.Empty{})
	if err != nil {
		t.Errorf("Fail to get cards %v", err)
	}

	if len(cards.Cards) != 0 {
		t.Errorf("Delete all has failed: %v", cards.Cards)
	}
}

func TestDelete(t *testing.T) {
	card := pb.Card{}
	card.Hash = "todelete"
	card2 := pb.Card{}
	card2.Hash = "toretain"
	s := InitTestServer(true)

	cardlist := pb.CardList{}
	cardlist.Cards = append(cardlist.Cards, &card)
	cardlist.Cards = append(cardlist.Cards, &card2)

	cards, err := s.AddCards(context.Background(), &cardlist)

	if err != nil {
		t.Errorf("Error in adding a card: %v", err)
	}

	if len(cards.Cards) != 2 {
		t.Errorf("Not enough cards: %v", len(cards.Cards))
	}

	deleteReq := pb.DeleteRequest{}
	deleteReq.Hash = "todelete"
	cards, _ = s.DeleteCards(context.Background(), &deleteReq)

	cards, err = s.GetCards(context.Background(), &pb.Empty{})
	if err != nil {
		t.Errorf("Error getting cards: %v", err)
	}

	if len(cards.Cards) != 1 {
		t.Errorf("Card has not been deleted correctly: %v:%v", len(cards.Cards), cards.Cards)
	}

	if cards.Cards[0].Hash != "toretain" {
		t.Errorf("Card has not been retained correctly: %v", cards.Cards)
	}
}

func TestRestart(t *testing.T) {
	card := pb.Card{Text: "What the", Hash: "hello"}
	s := InitTestServer(true)
	cardlist := pb.CardList{}
	cardlist.Cards = append(cardlist.Cards, &card)

	cards, err := s.AddCards(context.Background(), &cardlist)

	if err != nil {
		t.Errorf("Error adding card: %v", err)
	}

	if len(cards.Cards) != 1 {
		t.Errorf("Card has not beed added: %v", len(cards.Cards))
	}

	s2 := InitTestServer(false)
	cards, err = s2.GetCards(context.Background(), &pb.Empty{})
	if err != nil {
		t.Errorf("Error getting cards: %v", err)
	}

	if len(cards.Cards) != 1 {
		t.Errorf("Cards have not been returned: %v", cards)
	}
}

func TestRemoveStale(t *testing.T) {
	card := pb.Card{ExpirationDate: time.Now().Unix() - 10}
	s := InitTestServer(true)

	cardlist := pb.CardList{}
	cardlist.Cards = append(cardlist.Cards, &card)

	cards, err := s.AddCards(context.Background(), &cardlist)

	if err != nil {
		t.Errorf("Error adding card: %v", err)
	}

	if len(cards.Cards) != 1 {
		t.Errorf("Card has not beed added: %v", len(cards.Cards))
	}

	cards, err = s.GetCards(context.Background(), &pb.Empty{})
	if err != nil {
		t.Errorf("Error getting cards: %v", err)
	}

	if len(cards.Cards) != 0 {
		t.Errorf("Card has not been removed: %v:%v", len(cards.Cards), cards.Cards)
	}
}
