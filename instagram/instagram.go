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
     urlv:= "https://api.instagram.com/oauth/access_token"
     log.Printf("CODE = %v", code)
     resp, err := http.PostForm(urlv, url.Values{"client_id": {client}, "client_secret": {secret}, "grant_type":{"authorization_code"}, "redirect_uri": {"http://localhost:8090"}, "code": {code}, "scope": {"public_content+likes"}})
     if err != nil {
     	panic(err)
     }

     defer resp.Body.Close()
     body, _ := ioutil.ReadAll(resp.Body)

     return body
}

func ReadCards(client string, secret string) {
     log.Printf("Starting read")
     conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
     if err != nil {
       	  log.Fatalf("did not connect: %v", err)
     }
     defer conn.Close()
     clientReader := pb.NewCardServiceClient(conn)
     list, err := clientReader.GetCards(context.Background(), &pb.Empty{})

     log.Printf("READ = %v", list)

     access_code := make([]byte,0)
     for _, card := range(list.Cards) {
     	 log.Printf("READ %v", card)
     	 if card.Hash == "instagramauthresp" {
	    code := card.Text[11:43]
	    access_code = processAccessCode(client, secret, code)
	 }
     }

     //Write the access card out to a file
     if len(access_code) > 0 {
     err = ioutil.WriteFile("access_code", access_code, 0644)
     if err != nil {
     	panic(err)
     }
     }
}

func WriteAuthCard(client string) pb.CardList {
	authUrl := "https://api.instagram.com/oauth/authorize/?client_id=CLIENT-ID&redirect_uri=http://localhost:8090&response_type=code"
	newAuthUrl := strings.Replace(authUrl, "CLIENT-ID", client, 1)
	card := pb.Card{}
	card.Text = newAuthUrl
	card.Action = pb.Card_VISITURL
	card.Hash = "instagramauth"
	cards := pb.CardList{}
	cards.Cards = append(cards.Cards, &card)

	return cards
}

func WriteInstagramCards(user string, access_code string) pb.CardList{

     cards := pb.CardList{}
     
     urlv:= "https://api.instagram.com/v1/users/" + user + "/media/recent/?access_token=" + access_code + "&count=10"
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

	for _, pico := range(pics) {
	    pic := pico.(map[string]interface{})
	    captionObj := pic["caption"]
	    caption := ""
	    if captionObj != nil {
	       caption = captionObj.(map[string]interface{})["text"].(string)
	       }
	    var images map[string]interface{}
	    images = pic["images"].(map[string]interface{})
	    image := images["standard_resolution"].(map[string]interface{})["url"].(string)

	    card := pb.Card{}
	    card.Text = caption
	    card.Image = image
	    cards.Cards = append(cards.Cards, &card)
	}
	return cards
}

func main() {
	var clientId = flag.String("client", "", "Client ID for accessing Instagram")
	var secret = flag.String("secret", "", "Secret for accessing Instagram")
	flag.Parse()

	text, err := ioutil.ReadFile("access_code")

	if err != nil {
		cards := WriteAuthCard(*clientId)
		conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
		
		defer conn.Close()
		client := pb.NewCardServiceClient(conn)
		_, err = client.AddCards(context.Background(), &cards)
		if err != nil {
		   log.Printf("Problem adding cards %v", err)
		}

		log.Printf("Reading cards")
		ReadCards(*clientId, *secret)		
	} else {
	  var dat map[string]interface{}
	  if err := json.Unmarshal([]byte(text), &dat); err != nil {
	     panic(err)
	     }
	  cards := WriteInstagramCards("50987102", dat["access_token"].(string))
	  		conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
		
		defer conn.Close()
		client := pb.NewCardServiceClient(conn)
		_, err = client.AddCards(context.Background(), &cards)
		if err != nil {
		   log.Printf("Problem adding cards %v", err)
		}

	}
}
