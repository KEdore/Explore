package main

import (
	"flag"
	"log"
	"net"

	"github.com/KEdore/Explore/server"
	pb "github.com/KEdore/Explore/proto"
	"google.golang.org/grpc"
)

func main() {
	// Allow port override
	port := flag.String("port", "8080", "The server port")
	flag.Parse()

	lis, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalf("Failed to listen on port %s: %v", *port, err)
	}

	grpcServer := grpc.NewServer()
	store := server.NewInMemoryStore()
	exploreService := server.NewExploreServiceServer(store)
	pb.RegisterExploreServiceServer(grpcServer, exploreService)

	log.Printf("Server listening on port %s...", *port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
