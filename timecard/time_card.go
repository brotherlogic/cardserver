package main

import "log"
import "strconv"
import "time"
import "golang.org/x/net/context"
import "google.golang.org/grpc"
import pb "github.com/brotherlogic/cardserver/card"

func prepend(str string) string {
	if len(str) == 1 {
		return "0" + str
	}
	return str
}

func build() pb.CardList {
	// Generate 10 cards from the current time onwards
	now := time.Now().Truncate(time.Minute)
	cards := pb.CardList{}
	for i := 0; i < 10; i++ {
		card := pb.Card{}
		card.Text = prepend(strconv.Itoa(now.Hour())) + ":" + prepend(strconv.Itoa(now.Minute()))
		card.Hash = "timecard" + card.Text
		card.Priority = 1
		card.ApplicationDate = now.Unix()

		//Add a minute
		now = now.Add(time.Minute)
		card.ExpirationDate = now.Unix()

		cards.Cards = append(cards.Cards, &card)
	}

	return cards
}

func main() {
	cards := build()
	conn, err := grpc.Dial("localhost:50055", grpc.WithInsecure())

	defer conn.Close()
	client := pb.NewCardServiceClient(conn)
	_, err = client.AddCards(context.Background(), &cards)
	if err != nil {
		log.Printf("Problem adding cards %v", err)
	}
}
