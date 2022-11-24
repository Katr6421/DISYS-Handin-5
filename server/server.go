package main

import (
	"context"
	"time"

	//"database/sql/driver"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	proto "simpleGuide/grpc"
	"strconv"
)

// Struct for the Server to store to keep track of the clients
type ClientStream struct {
	name     string
	clientID int
	stream   *proto.TimeAsk_ConnectToServerServer
	chQuit   chan int
}

// Struct that will be used to represent the Server.
type Server struct {
	proto.UnimplementedAuctionServer // Necessary
	id                               int32
	clients                          []*ClientStream
	servers                          map[int32]proto.AuctionClient
	ctx                              context.Context
	isLeader                         bool
	isDead                           bool
	chDeadOrAlive                    chan int32
}

// Sets the serverport to 5454
var port = flag.Int("port", 5454, "server port number")


func main() {
	// Log to custom file
	LOG_FILE := "../custom.log"
	// Open log file - or create it, if it doesn't exist
	logFile, err := os.OpenFile(LOG_FILE, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer logFile.Close()

	// Set log out put
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
		id:      ownPort,
		clients: []*ClientStream{},
		servers: make(map[int32]proto.AuctionClient),
		ctx:     ctx,
	}

	// Start the server
	go startServer(server, ownPort)

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
	}

	if server.isLeader {
		for {
			err, failingServerId := server.SendingHeartbeat()

			// If SendingHeartbeat returns an error, it means one of the backup replicas didn't respond
			if err != nil{
				log.Printf("Primary replica didn't get a response from backup replica %d. This backup replica will be removed", failingServerId)
				delete(server.servers, failingServerId) // Removes non-responding replica from map of connected servers
			}

			time.Sleep(15 * time.Second)  


			//tjekker channel
			if len(server.chDeadOrAlive) == len(server.servers) {
				//leaderen er død
				//elect ny leader
			}

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

func (s *Server) SendingHeartbeat() (Error error, failingServerId int32) {

	var requiredResponses int = 2; // Number of responses if all backups responds
	var numberOfResponses int = 0; 
	var nonRespondingReplica int32 = 0; 
	var error error = nil; 

	BipBip := &proto.SendHeartbeat{
		ServerID: s.id,
	}

	log.Printf(" !!! Primary replica %d sends heartbeat to all backup replicas !!!", s.id)
	
	for id, server := range s.servers {
		response, err := server.Heartbeat(s.ctx, BipBip) 

		// If Heartbeat() returns an error, the backup replica didn't respond
		if err != nil {
			error = err; 
			nonRespondingReplica = id; 
			requiredResponses--; // Non-responding replica will be removed, so required responses is one lower
		}

		// Backup replica responded
		if err == nil {
		numberOfResponses++;
		log.Printf("Primary replica %d got the response: %d from backup replica %d", s.id, response.Ack, response.ServerID)
		}
	}

	// If primary didn't get responses from every backup
	if (requiredResponses != numberOfResponses) {
		return error, nonRespondingReplica
	}

	// If primary did get responses from every backup
	return nil, 0;
}

func (s *Server) Heartbeat(ctx context.Context, bipbip *proto.SendHeartbeat) (*proto.Response, error) {
	//når de modtager et heartbeat

	ack := &proto.Response{
		Ack:      1,
		ServerID: s.id, //id på den backup server der svarer
	}

	log.Printf("Backup replica %d sends ack to heartbeat", s.id)
	return ack, nil
}

/*
func (s *Server) SendMessage(ctx context.Context, msg *proto.ClientPublishMessage) (*proto.ServerPublishMessageOk, error) {
	//update LamportTime. compare Servers' time and the time for the msg we received
	if msg.LamportTimestamp > s.LamportTimestamp {
		s.LamportTimestamp = msg.LamportTimestamp + 1
	} else {
		s.LamportTimestamp += 1
	}
	log.Printf("Received a message from participant %d at Lamport time %d\n", msg.ClientId, s.LamportTimestamp)

	if msg.Message == "quit" {
		for i, client := range s.clients {
			if msg.ClientId == int64(client.clientID) {
				//removes the client who quited - the same as pop. '...' means we want to add to our array
				s.clients = append(s.clients[:i], s.clients[i+1:]...)

				if len(s.clients) < 2 {
					log.Printf("There is %d client connected", len(s.clients))
				} else {
					log.Printf("There are %d clients connected", len(s.clients))
				}

				//sends a message that we want to break the connection to the server
				client.chQuit <- 0
				s.SendToAllClients(fmt.Sprintf("Participant %s left Chitty-Chat at Server Lamport time %d", client.name, s.LamportTimestamp))
				break
			}
		}
	} else {
		s.SendToAllClients(msg.Message)
	}

	return &proto.ServerPublishMessageOk{
		Time:       time.Now().String(),
		ServerName: s.name,
	}, nil
}*/
/*
func (s *Server) SendToAllClients(msg string) {
	//for each client send them a stream of message
	//_, _ = ignorerer variablerne i metoden
	for _, client := range s.clients {
		log.Printf("Sending the message to participant %s with id: %d\n", client.name, client.clientID)
		(*client.stream).Send(&proto.MessageStreamConnection{
			StreamMessage:    msg,
			LamportTimestamp: s.LamportTimestamp,
		})
	}
}*/

/*
func (s *Server) ConnectToServer(msg *proto.ClientConnectMessage, stream proto.TimeAsk_ConnectToServerServer) error {
	log.Printf("%s connected to the server at Lamport time %d\n", msg.Name, s.LamportTimestamp)

	clientStream := &ClientStream {
		name: msg.Name,
		clientID: int(msg.ClientId),
		stream: &stream,
		chQuit: make(chan int),
	}

	//saves the clients/participants that are connected to the Server
	s.clients = append(s.clients, clientStream)

	//sends message to all the connected clients
	s.SendToAllClients(fmt.Sprintf("Participant %s joined Chitty-Chat at Server Lamport time %d", msg.Name, s.LamportTimestamp))

	if len(s.clients) < 2 {
		log.Printf("There is %d client connected", len(s.clients))
	} else {
		log.Printf("There are %d clients connected", len(s.clients))
	}

	//as long as there is no message, the channel will stay open
	<-clientStream.chQuit

	return nil
}*/

func (s *Server) Bid(ctx context.Context, amount *proto.Amount) (*proto.ConfirmationOfBid, error) {
	//TODO implement me
	panic(recover())
}

func (s *Server) Result(ctx context.Context, request *proto.Request) (*proto.AuctionStatus, error) {
	//TODO implement me
	panic(recover())
}
