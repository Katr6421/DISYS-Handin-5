package main

import (
	"context"
	"flag"
	"google.golang.org/grpc"
	"log"
	"net"
	proto "simpleGuide/grpc"
	"strconv"
	"time"
	"fmt"
)

//Struct for the Server to store to keep track of the clients
type ClientStream struct {
	name string
	clientID int
	stream *proto.TimeAsk_ConnectToServerServer
	chQuit chan int
}

// Struct that will be used to represent the Server.
type Server struct {
	proto.UnimplementedTimeAskServer // Necessary
	name                             string
	port                             int
	clients                          []*ClientStream
	LamportTimestamp                 int64
}

// Sets the serverport to 5454
var port = flag.Int("port", 5454, "server port number")

func main() {
	// This parses the flag and sets the correct/given corresponding values.
	flag.Parse()

	// Create a server struct
	server := &Server {
		name: "serverName",
		port: *port,
		clients: []*ClientStream{},
		LamportTimestamp: 1,
	}

	// Start the server
	go startServer(server)

	// Keep the server running until it is manually quit
	for {

	}
}

func startServer(server *Server) {

	// Create a new grpc server
	grpcServer := grpc.NewServer()

	// Make the server listen at the given port (convert int port to string)
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(server.port))
	if err != nil {
		log.Fatalf("Could not create the server %v", err)
	}
	log.Printf("Started server at port: %d\n", server.port)

	// Register the grpc server and serve its listener
	proto.RegisterTimeAskServer(grpcServer, server)
	serveError := grpcServer.Serve(listener)
	if serveError != nil {
		log.Fatalf("Could not serve listener")
	}
}

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
}


func (s *Server) SendToAllClients(msg string) {
	//for each client send them a stream of message
	//_, _ = ignorerer variablerne i metoden
	for _, client := range s.clients {
		log.Printf("Sending the message to participant %s with id: %d\n", client.name, client.clientID)
		(*client.stream).Send(&proto.MessageStreamConnection {
			StreamMessage: msg,
			LamportTimestamp: s.LamportTimestamp,
		})
	}
}

func (s *Server) ConnectToServer(msg *proto.ClientConnectMessage, stream proto.TimeAsk_ConnectToServerServer) error {
	s.LamportTimestamp += 1
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
}



