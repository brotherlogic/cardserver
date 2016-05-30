package main

import (
       pb "github.com/brotherlogic/cardserver/card"
       "strings"
       "testing"
)

func TestRespondToInstagramRating(t *testing.T) {
     card1 := pb.Card {
     	   Action: pb.Card_RATING,
	   Text: "picture_id",
	   Hash: "instagramrating",
	   ActionMetadata: []string{"1"},
     }

     url := ProcessInstagramRating(card1)

     if !strings.Contains(url, "picture_id") {
     	t.Errorf("Processing Instagram rating leads to bad URL: %v", url)
     }
}