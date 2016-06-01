package main

import "encoding/json"
import "flag"
import "io/ioutil"
import "log"
import "net/http"
import "net/url"
import "strings"
import "golang.org/x/net/context"
import "google.golang.org/grpc"
import pb "github.com/brotherlogic/cardserver/card"

func processAccessCode(client string, secret string, code string) []byte {
	urlv := "https://api.instagram.com/oauth/access_token"
	resp, err := http.PostForm(urlv, url.Values{"client_id": {client}, "client_secret": {secret}, "grant_type": {"authorization_code"}, "redirect_uri": {"http://localhost:8090"}, "code": {code}, "scope": {"public_content+likes"}})
	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	return body
}

func processInstagramRating(card pb.Card) string {
	pictureID := card.Text
	urlv := "https://api.instagram.com/v1/media/" + pictureID + "/likes"
	return urlv
}

func visitURL(urlv string, accessToken string) []byte {
	resp, err := http.PostForm(urlv, url.Values{"access_token": {accessToken}})

	if err != nil {
		panic(err)
	}

	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	return body
}

func readCards(client string, secret string) {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	clientReader := pb.NewCardServiceClient(conn)
	list, err := clientReader.GetCards(context.Background(), &pb.Empty{})

	var accessCode []byte
	for _, card := range list.Cards {
		if card.Hash == "instagramauthresp" {
			code := card.Text[11:43]
			accessCode = processAccessCode(client, secret, code)
		}
	}

	//Write the access card out to a file
	if len(accessCode) > 0 {
		err = ioutil.WriteFile("access_code", accessCode, 0644)
		if err != nil {
			panic(err)
		}

		conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())

		defer conn.Close()
		client := pb.NewCardServiceClient(conn)
		req := pb.DeleteRequest{Hash: "instagramauthresp"}
		_, err = client.DeleteCards(context.Background(), &req)
		if err != nil {
			log.Printf("Problem delete auth resp card %v", err)
		}

	}
}

func writeAuthCard(client string) pb.CardList {
	authURL := "https://api.instagram.com/oauth/authorize/?client_id=CLIENT-ID&redirect_uri=http://localhost:8090&response_type=code"
	newAuthURL := strings.Replace(authURL, "CLIENT-ID", client, 1)
	card := pb.Card{}
	card.Text = newAuthURL
	card.Action = pb.Card_VISITURL
	card.Hash = "instagramauth"
	cards := pb.CardList{}
	cards.Cards = append(cards.Cards, &card)

	return cards
}

func isImageLiked(mediaID string, accessToken string) bool {
	likesURL := "https://api.instagram.com/v1/media/" + mediaID + "/likes?access_token=" + accessToken
	resp, err := http.Get(likesURL)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var dat map[string]interface{}
	err = json.Unmarshal([]byte(body), &dat)
	if err != nil {
		panic(nil)
	}

	var likes []interface{}
	likes = dat["data"].([]interface{})
	for _, persono := range likes {
		person := persono.(map[string]interface{})
		if person["username"].(string) == "brotherlogic" {
			return true
		}
	}

	return false
}

func writeInstagramCards(user string, accessCode string) pb.CardList {

	cards := pb.CardList{}

	urlv := "https://api.instagram.com/v1/users/" + user + "/media/recent/?access_token=" + accessCode + "&count=10"
	resp, err := http.Get(urlv)
	if err != nil {
		panic(nil)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(nil)
	}
	var dat map[string]interface{}
	err = json.Unmarshal([]byte(body), &dat)
	if err != nil {
		panic(nil)
	}

	var pics []interface{}
	pics = dat["data"].([]interface{})

	for _, pico := range pics {
		pic := pico.(map[string]interface{})
		captionObj := pic["caption"]
		caption := ""
		if captionObj != nil {
			caption = captionObj.(map[string]interface{})["text"].(string)
		}
		var images map[string]interface{}
		images = pic["images"].(map[string]interface{})
		image := images["standard_resolution"].(map[string]interface{})["url"].(string)

		liked := isImageLiked(pic["id"].(string), accessCode)

		if !liked {
			card := pb.Card{}
			card.Text = caption
			card.Image = image
			card.Priority = 20
			card.Hash = pic["id"].(string)
			cards.Cards = append(cards.Cards, &card)
		}
	}
	return cards
}

func main() {
	var clientID = flag.String("client", "", "Client ID for accessing Instagram")
	var secret = flag.String("secret", "", "Secret for accessing Instagram")
	flag.Parse()
	text, err := ioutil.ReadFile("access_code")

	if err != nil {
		cards := writeAuthCard(*clientID)
		conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())

		defer conn.Close()
		client := pb.NewCardServiceClient(conn)
		_, err = client.AddCards(context.Background(), &cards)
		if err != nil {
			log.Printf("Problem adding cards %v", err)
		}

		readCards(*clientID, *secret)
	} else {
		var dat map[string]interface{}
		if err := json.Unmarshal([]byte(text), &dat); err != nil {
			log.Printf("Error unmarshalling %v with %v and %v", string(text), *clientID, *secret)
			panic(err)
		}
		if dat == nil || dat["access_token"] == nil {
			log.Printf("Cannot get access token: %v from %v given $v,%v", dat, string(text), *clientID, *secret)
		}
		cards := writeInstagramCards("50987102", dat["access_token"].(string))
		conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())

		defer conn.Close()
		client := pb.NewCardServiceClient(conn)
		_, err = client.AddCards(context.Background(), &cards)
		if err != nil {
			log.Printf("Problem adding cards %v", err)
		}

	}
}
