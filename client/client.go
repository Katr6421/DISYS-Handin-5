package main

import (
	"bufio"
	"context"
	"flag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	//"io"
	"log"
	"os"
	proto "simpleGuide/grpc"
	"strconv"
)

type Client struct {
	id               int
	name 			 string
	//stream           *proto.TimeAsk_ConnectToServerClient
}

// go run . -name Hannah. Command to connect to server via a chosen name.
var (
	name       = flag.String("name", "<name>", "Name of this participant")
	serverPort = flag.Int("sPort", 8080, "server port number (should match the port used for the server)")
)

func main() {
	// Log to custom file
	LOG_FILE := "../custom.log"
	// Open log file - or create it, if it doesn't exist
	logFile, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer logFile.Close()
		
	// Set log output
	log.SetOutput(logFile)

	// Log date-time and filename
	log.SetFlags(log.Lshortfile | log.LstdFlags)
	

	// Parse the flags to get the port for the client
	flag.Parse()

	
	// Create a client
	client := &Client{
		id:         1,
		name:       *name,
	}


	go waitForBid(client)

	for{

	}

}

func waitForBid(client *Client){
	serverConnection, _ := connectToServer(client)

	// Wait for bid in the client terminal
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()

		// Check if inout is an integer and convert
		bid, err := strconv.Atoi(input)
		if err != nil {
			log.Printf("%v is not a valid input. Please write a number", input)
		}
		if err == nil {
		log.Printf("Client %d made bid: %v\n", client.id, input)

		// Ask the server for the time
		bidReturnMessage, err := serverConnection.Bid(context.Background(), &proto.Amount{
			ClientId: int32(client.id),
			Bid: 	  int32(bid),
		})

		if err != nil {
			log.Printf(err.Error()) // hvis client ikke får respons, skal ny leder vælges
		} else {
			log.Printf("Server %v confirms with message: %s\n", bidReturnMessage.ServerId, bidReturnMessage.ConfirmationMsg)
		}
	}
	}

}


func connectToServer(client *Client) (proto.AuctionClient, error) {
	// Dial the server at the specified port.
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(*serverPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Client %d could not connect to port %d", client.id, *serverPort)
	} else {
		log.Printf("Client %d connected to the server at port %d\n", client.id, *serverPort)
	}
	return proto.NewAuctionClient(conn), nil
}
