package main

import "flag"
import "log"
import "strings"
import "golang.org/x/net/context"
import "google.golang.org/grpc"
import pb "github.com/brotherlogic/cardserver/card"

func WriteAuthCard(client string) pb.CardList {
	authUrl := "https://api.instagram.com/oauth/authorize/?client_id=CLIENT-ID&redirect_uri=http://localhost:8090/&response_type=token"
	newAuthUrl := strings.Replace(authUrl, "CLIENT-ID", client, 0)

	card := pb.Card{}
	card.Text = newAuthUrl
	cards := pb.CardList{}
	cards.Cards = append(cards.Cards, &card)

	return cards
}

func main() {
	var clientId = flag.String("client", "", "Client ID for accessing Instagram")
	cards := WriteAuthCard(*clientId)
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())

	defer conn.Close()
	client := pb.NewCardServiceClient(conn)
	_, err = client.AddCards(context.Background(), &cards)
	if err != nil {
		log.Printf("Problem adding cards %v", err)
	}
}
