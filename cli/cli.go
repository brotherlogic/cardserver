package main

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"golang.org/x/net/context"
	"google.golang.org/grpc"

	pb "github.com/brotherlogic/cardserver/card"
	pbdi "github.com/brotherlogic/discovery/proto"
)

func findServer(name string) (string, int) {
	conn, err := grpc.Dial("192.168.86.42:50055", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Cannot reach discover server: %v (trying to discover %v)", err, name)
	}
	defer conn.Close()

	registry := pbdi.NewDiscoveryServiceClient(conn)
	r, err := registry.Discover(context.Background(), &pbdi.RegistryEntry{Name: name})

	if err != nil {
		log.Printf("Failure to list: %v", err)
		return "", -1
	}
	return r.Ip, int(r.Port)
}

func main() {

	if len(os.Args) <= 1 {
		fmt.Printf("Commands: build run\n")
	} else {
		switch os.Args[1] {
		case "clear":
			host, port := findServer("cardserver")

			conn, _ := grpc.Dial(host+":"+strconv.Itoa(port), grpc.WithInsecure())
			defer conn.Close()

			registry := pb.NewCardServiceClient(conn)
			_, err := registry.DeleteCards(context.Background(), &pb.DeleteRequest{HashPrefix: "disco"})
			if err != nil {
				log.Printf("Error deleting cards: %v", err)
			}
		case "list":
			host, port := findServer("cardserver")

			if port > 0 {

				conn, _ := grpc.Dial(host+":"+strconv.Itoa(port), grpc.WithInsecure())
				defer conn.Close()

				registry := pb.NewCardServiceClient(conn)
				rs, err := registry.GetCards(context.Background(), &pb.Empty{})
				if err != nil {
					log.Printf("Error deleting cards: %v", err)
				}
				for _, res := range rs.Cards {
					log.Printf("CARD: %v", res)
				}
			}
		}
	}
}
