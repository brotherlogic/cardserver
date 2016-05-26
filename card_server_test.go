package main

import (
	pb "github.com/brotherlogic/cardserver/card"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"net"
	"testing"
	"time"
)

func TestAdd(t *testing.T) {
	card := pb.Card{}
	s := InitServer()

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

func TestRemoveStale(t *testing.T) {
	card := pb.Card{ExpirationDate: time.Now().Unix() - 10}
	s := InitServer()

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

func TestRunServer(t *testing.T) {
	go func() {
		lis, err := net.Listen("tcp", ":50051")
		if err != nil {
			t.Errorf("Error opening port up")
		}
		s := grpc.NewServer()
		server := InitServer()
		pb.RegisterCardServiceServer(s, &server)
		s.Serve(lis)
	}()

	go func() {
		time.Sleep(5 * time.Second)
		conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
		if err != nil {
			t.Errorf("Error connecting to port")
		}
		defer conn.Close()
		client := pb.NewCardServiceClient(conn)

		card := pb.Card{}
		card.Text = "Testing"
		cardlist := pb.CardList{}
		cardlist.Cards = append(cardlist.Cards, &card)

		r, err := client.AddCards(context.Background(), &cardlist)
		if err != nil {
			t.Errorf("Error adding cards: %v", err)
		}

		r, err = client.GetCards(context.Background(), &pb.Empty{})
		if err != nil {
			t.Errorf("Error getting cards: %v", err)
		}
		if len(r.Cards) != 1 {
			t.Errorf("Not enough cards: %v", r)
		}
		if r.Cards[0].Text != "Testing" {
			t.Errorf("Card read is wrong: %v", r)
		}
	}()

	time.Sleep(10 * time.Second)
}
