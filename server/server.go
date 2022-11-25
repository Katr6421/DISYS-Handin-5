package main

import (
	"context"
	"strings"
	"time"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	proto "simpleGuide/grpc"
	"strconv"
	"google.golang.org/grpc"
	"sync"
)

type Auction struct{
	highestBid 		int32
	//bidders 	    []*int32
	highestBidderId int32
	auctionFinished bool	
}

// Struct for the Server to store to keep track of the clients
type ClientStream struct {
	name     		string
	clientID 		int32
	//stream   		*proto.TimeAsk_ConnectToServerServer
	//chQuit  		chan int
}

// Struct that will be used to represent the Server.
type Server struct {
	proto.UnimplementedAuctionServer // Necessary
	id                               int32
	clients                          []*ClientStream
	servers                          map[int32]proto.AuctionClient
	ctx                              context.Context
	isLeader                         bool
	auctionInfo     				 *Auction
	//isDead                           bool
	//chDeadOrAlive                    chan int32
}

// Sets the serverport to 5454
var port = flag.Int("port", 5454, "server port number")
var mu sync.Mutex
//var Leader int

func main() {
	// Log to custom file
	LOG_FILE := "../log.log"
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


	// This parses the flag and sets the correct/given corresponding values.
	//flag.Parse()

	arg1, _ := strconv.ParseInt(os.Args[1], 10, 32)
	ownPort := int32(arg1) + 8080
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create a server struct
	server := &Server{
		id:      	 ownPort,
		clients: 	 []*ClientStream{},
		servers: 	 make(map[int32]proto.AuctionClient),
		ctx:     	 ctx,
		auctionInfo: new(Auction),
	}

	// Start the server
	go startServer(server, ownPort)

	// End auction after 5 minutes
	time.Sleep(5 * time.Minute)  

	// Leader ends auction
	if server.isLeader{
		server.auctionInfo.auctionFinished = true
		
		// Update the other replicas to know auction is finished
		Update := &proto.AuctionUpdate{
			ServerId: 			server.id,
			HighestBid: 		server.auctionInfo.highestBid,
			HiggestBidderId: 	server.auctionInfo.highestBidderId,
			AuctionStatus: 		false,
		}

		for id, backupServer := range server.servers{
			_, err := backupServer.UpdateBackup(server.ctx, Update) 

			// If UpdateBackup() returns an error, the backup replica didn't respond
			if err != nil {
				log.Printf("Couldn't update backup replica %d. This backup replica will be removed", id)
				delete(server.servers, id) // Removes non-responding replica from map of connected servers
			}
		}
	
	log.Printf("Auction is finished. Winner was %d with a bid of %d kr", server.auctionInfo.highestBidderId, server.auctionInfo.highestBid)
	}

	// Keep the server running until it is manually quit
	for {

	}
}

func startServer(server *Server, ownPort int32) {

	// Create listener tcp on port ownPort
	list, err := net.Listen("tcp", fmt.Sprintf(":%v", ownPort))
	if err != nil {
		log.Fatalf("Failed to listen on port: %v", err)
	}
	log.Printf("Started server at port: %d\n", server.id)

	grpcServer := grpc.NewServer()

	// Register the grpc server and serve its listener
	proto.RegisterAuctionServer(grpcServer, server)

	go func() {
		if err := grpcServer.Serve(list); err != nil {
			log.Fatalf("failed to server %v", err)
		}
	}()

	// Dials three servers
	for i := 0; i < 3; i++ {
		port := int32(8080) + int32(i)

		if port == ownPort {
			continue
		}

		var conn *grpc.ClientConn
		fmt.Printf("Trying to dial: %v\n", port)
		conn, err := grpc.Dial(fmt.Sprintf(":%v", port), grpc.WithInsecure(), grpc.WithBlock())
		if err != nil {
			log.Fatalf("Could not connect: %s", err)
		}
		defer conn.Close()
		c := proto.NewAuctionClient(conn)
		server.servers[port] = c
		log.Printf("Connected to server: %v\n", port)
	}

	if server.id == 8080 {
		server.isLeader = true // Primary replica is 8080 by default 
		//Leader = 8080
	}

	if server.isLeader {
		for {
			failingServerId, err := server.SendingHeartbeat()

			// If SendingHeartbeat returns an error, it means one of the backup replicas didn't respond
			if err != nil{
				log.Printf("Primary replica didn't get a heartbeat-response from backup replica %d. This backup replica will be removed", failingServerId)
				delete(server.servers, failingServerId) // Removes non-responding replica from map of connected servers
			}

			time.Sleep(15 * time.Second)  


			//tjekker channel
			//if len(server.chDeadOrAlive) == len(server.servers) {
				//leaderen er død
				//elect ny leader
			//}

		}
	}


	//if !server.isLeader {
	//	time.Sleep(30 * time.Second)
	////	log.Printf("hej")
		//if bipbip.isDead == true {
		//election
		//}

	//}
	//server på index 0 - kalder Heartbeat for X sek
	//servers på index 1++ - kalder Response hvert x sek
	//sæt if statement på hvis de andre servers venter for længe -> vælg ny leader
	////den nye leder er index 0+1 og resten er så fra index 1+1

}


//// METHODS REGARDING HEARTBEAT
func (s *Server) SendingHeartbeat() (failingServerId int32, Error error) {

	var requiredResponses int = 2; // Number of responses if all backups responds
	var numberOfResponses int = 0; 
	var nonRespondingReplica int32 = 0; 
	var error error = nil; 

	BipBip := &proto.SendHeartbeat{
		ServerId: s.id,
	}

	log.Printf(" !!! Primary replica %d sends heartbeat to all backup replicas !!!", s.id)
	
	for id, server := range s.servers {
		_, err := server.Heartbeat(s.ctx, BipBip) 

		// If Heartbeat() returns an error, the backup replica didn't respond
		if err != nil {
			error = err; 
			nonRespondingReplica = id; 
			requiredResponses--; // Non-responding replica will be removed, so required responses is one lower
		}

		// Backup replica responded
		if err == nil {
		numberOfResponses++;
		//log.Printf("Primary replica %d got the response: %d from backup replica %d", s.id, response.Ack, response.ServerId)
		}
	}

	// If primary didn't get responses from every backup
	if (requiredResponses != numberOfResponses) {
		return nonRespondingReplica, error
	}

	// If primary did get responses from every backup
	return 0, nil;
}

func (s *Server) Heartbeat(ctx context.Context, bipbip *proto.SendHeartbeat) (*proto.ResponseToHeartbeat, error) {
	//når de modtager et heartbeat

	ack := &proto.ResponseToHeartbeat{
		ServerId: s.id, //id på den backup server der svarer
	}

	log.Printf("Backup replica %d sends ack to heartbeat", s.id)
	return ack, nil
}



//// METHODS REGARDING BIDDING
func (s *Server) Bid(ctx context.Context, amount *proto.BidAmount) (*proto.ConfirmationOfBid, error) {

	log.Printf("Client with ID %d made bid: %d kr\n", amount.ClientId, amount.Bid)
	
	// If auction is open
	if (!s.auctionInfo.auctionFinished){
		// If bid is highest
		if (amount.Bid > s.auctionInfo.highestBid){
			
			mu.Lock() //Lock so changes cannot be made before every replica is updated

			// Update own auction
			s.auctionInfo.highestBid = amount.Bid
			s.auctionInfo.highestBidderId = amount.ClientId

			// Update the other replicas 
			Update := &proto.AuctionUpdate{
				ServerId: 			s.id,
				HighestBid: 		amount.Bid,
				HiggestBidderId: 	amount.ClientId,
				AuctionStatus:      s.auctionInfo.auctionFinished,
			}

			for id, backupServer := range s.servers{
				_, err := backupServer.UpdateBackup(s.ctx, Update) 

				// If UpdateBackup() returns an error, the backup replica didn't respond
				if err != nil {
					log.Printf("Couldn't update backup replica %d. This backup replica will be removed", id)
					delete(s.servers, id) // Removes non-responding replica from map of connected servers
				}

				// Backup replica responded
				if err == nil {
					//log.Printf("Backup replica %d responded to update request", id)
				}

				defer mu.Unlock()
			}
			
			// Print
			log.Printf("Client %d now has highest bid of %d kr", amount.ClientId, amount.Bid)

			// Answer client
			return &proto.ConfirmationOfBid{
				ServerId: 	s.id, 
				ConfirmationMsg:  "Accepted",
			}, nil

		}

		// If bid is not higher
		if (amount.Bid <= s.auctionInfo.highestBid){

			// Print
			log.Printf("Client %d bid %d kr, but it was not higher than the current highest bid of %d kr", amount.ClientId, amount.Bid, s.auctionInfo.highestBid)

			//Answer client
			return &proto.ConfirmationOfBid{
				ServerId: 	s.id, 
				ConfirmationMsg:  "Rejected",
			}, nil
		}
	}

	//If auction is closed 
	if (s.auctionInfo.auctionFinished){

		// Print
		log.Printf("Client %d bid %d kr, but auction is closed", amount.ClientId, amount.Bid)

		// Answer client
		return &proto.ConfirmationOfBid{
			ServerId: 	s.id, 
			ConfirmationMsg:  "Rejected",
		}, nil
	}

	return nil, nil
}

func (s *Server) UpdateBackup(ctx context.Context, msg *proto.AuctionUpdate) (*proto.ConfirmationOfUpdate, error){	
	// Update auctionInfo
	s.auctionInfo.highestBid = msg.HighestBid;
	s.auctionInfo.highestBidderId = msg.HiggestBidderId

	log.Printf("Backup replica %d has updated its auction information", s.id)

	// Send answer
	ack := &proto.ConfirmationOfUpdate{
		ServerId: s.id, 
	}

	return ack, nil
}


//// METHODS REGARDING STATUS OF AUCTION
func (s *Server) Result(ctx context.Context, request *proto.RequestStatus) (*proto.AuctionStatus, error) {
	// Answer client's request
	log.Printf("Client %d asked what the status of the auction is", request.ClientId)

	// Builds the status message depending on whether the auction is open or closed
	builder := strings.Builder{}    
	builder.WriteString("Status of action: ")
	if s.auctionInfo.auctionFinished{
		builder.WriteString(fmt.Sprintf("Auction is over. Winning bid was %d kr ", s.auctionInfo.highestBid))
	}  
	if !s.auctionInfo.auctionFinished {
		builder.WriteString(fmt.Sprintf("Auction is still open. Current highest bid is %d kr ", s.auctionInfo.highestBid))
	}
	builder.WriteString(fmt.Sprintf("by client %d", s.auctionInfo.highestBidderId))
    msg := builder.String() 

	// Send answer
	response := &proto.AuctionStatus{
		ServerId: s.id,
		StatusMsg: msg,
	}

	return response, nil
}
