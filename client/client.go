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
	id   int
	name string
	//stream           *proto.TimeAsk_ConnectToServerClient
}

var (
	leaderId int32 = 8080
	serverPort = flag.Int("sPort", int(leaderId), "server port number (should match the port used for the server)")
)

func main() {
	// Log to custom file
	LOG_FILE := "../log2.log"
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
		id: os.Getpid(),
	}

	serverConnection, _ := connectToServer(client)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		input := scanner.Text()

		// If client wants to know status of the auction
		if input == "status" {
			go client.requestStatus(serverConnection)
		}

		// If client wants to make a bid
		if input != "status" {

			bid, err := strconv.Atoi(input)

			// If client inputted text instead of a number (unvalid bid)
			if err != nil {
				log.Printf("%v is not a valid input. Please write a number", input)
			}

			// If client inputted a number (valid bid)
			if err == nil {
				go client.makeABid(bid, serverConnection)
			}
		}
	}

	for {

	}
}

func (client *Client) makeABid(bid int, serverConnection proto.AuctionClient) {

	log.Printf("Client %d wants to make a bid: %v kr\n", client.id, bid)

	bidReturnMessage, err := serverConnection.Bid(context.Background(), &proto.BidAmount{
		ClientId: int32(client.id),
		Bid:      int32(bid),
	})

	if err != nil {
		log.Printf("Cannot connect to leader, trying to connect to another server") // hvis client ikke får respons, skal ny leder vælges?

		// Hvis 8080 er død, er 8082 den nye leader
		leaderId = 8082
		_, err := connectToServer(client)
		if (err != nil){
			// Hvis 8080 og 8082 er døde, er 8081 den nye leader
			leaderId = 8081
			_, err2 := connectToServer(client)
			if (err2 != nil){
				log.Printf("Can not connect to any server. They are all dead :(")
			} else{
				log.Printf("Client connected to server 8081")
			}
		} else {
			log.Printf("Client connected to server 8082")
		}
	} else {
		log.Printf("Server %v answers bid with message: %s\n", bidReturnMessage.ServerId, bidReturnMessage.ConfirmationMsg)
	}
}

func (client *Client) requestStatus(serverConnection proto.AuctionClient) {

	requestStatusMessage, err := serverConnection.Result(context.Background(), &proto.RequestStatus{
		ClientId: int32(client.id),
	})

	if err != nil {
			// Hvis 8080 er død, er 8082 den nye leader
			leaderId = 8082
			_, err := connectToServer(client)
			if (err != nil){
				// Hvis 8080 og 8082 er døde, er 8081 den nye leader
				leaderId = 8081
				_, err2 := connectToServer(client)
				if (err2 != nil){
					log.Printf("Can not connect to any server. They are all dead :(")
				} else{
					log.Printf("Client connected to server 8081")
				}
			} else {
				log.Printf("Client connected to server 8082")
			}
	} else {
		// Got response 
		log.Printf("Client %d got response from server %d. Message: %v", client.id, requestStatusMessage.ServerId, requestStatusMessage.StatusMsg)
	}
}

func connectToServer(client *Client) (proto.AuctionClient, error) {

	// Dial the server at the specified port.
	conn, err := grpc.Dial("localhost:"+strconv.Itoa(*serverPort), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Client %d could not connect to port %d", client.id, *serverPort)
		return nil, err
	} else {
		log.Printf("Client %d connected to the server at port %d\n", client.id, *serverPort)
	}
	return proto.NewAuctionClient(conn), nil
}
