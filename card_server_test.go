package main

import "testing"
import "golang.org/x/net/context"
import pb "github.com/brotherlogic/cardserver/card"

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