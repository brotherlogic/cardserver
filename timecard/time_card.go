package main

import "log"
import "strconv"
import "time"
import "golang.org/x/net/context"
import "google.golang.org/grpc"
import pb "github.com/brotherlogic/cardserver/card"

func Build() pb.CardList {
     // Generate 10 cards from the current time onwards
     now := time.Now().Truncate(time.Minute)
     cards := pb.CardList{}
     for i:=0; i < 10 ; i++ {
          card := pb.Card{}
	  card.Text = strconv.Itoa(now.Hour()) + ":" + strconv.Itoa(now.Minute())
	  card.ApplicationDate = now.Unix()

	  //Add a minute
	  now = now.Add(time.Minute)
	  card.ExpirationDate = now.Unix()

	  cards.Cards = append(cards.Cards, &card)
     }

     return cards
}

func main() {
     cards := Build()
     conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())

     defer conn.Close()
     client := pb.NewCardServiceClient(conn)
     _, err = client.AddCards(context.Background(), &cards)
     if err != nil {
     	log.Printf("Problem adding cards %v", err)
     }
}