// RunServer starts the gRPC server and returns a function to gracefully stop it.
package server

import (
	"context"
	"net"
	"log"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/KEdore/explore/internal/config"
	"github.com/KEdore/explore/internal/db"
	pb "github.com/KEdore/explore/proto"
	"github.com/KEdore/explore/internal/service"
)

// RunServer starts a gRPC server with the provided configuration and returns a function to stop the server.
// It initializes a MySQL database client and sets up the gRPC server to listen on the specified address.
func RunServer(ctx context.Context, cfg *config.Config) (stopFunc func(), err error) {
	database, err := db.NewMySQLClient(cfg.DBUser, cfg.DBPass, cfg.DBHost, cfg.DBName)
	if err != nil {
		return nil, err
	}
	// Make sure the database is closed when the server stops.
	// (You might want to handle this in your shutdown logic.)

	lis, err := net.Listen("tcp", cfg.ServerAddress)
	if err != nil {
		return nil, err
	}

	grpcServer := grpc.NewServer()
	pb.RegisterExploreServiceServer(grpcServer, service.NewExploreServer(database))
	reflection.Register(grpcServer)

	// Run the server in a goroutine.
	go func() {
		log.Printf("Server listening at %v", lis.Addr())
		if err := grpcServer.Serve(lis); err != nil {
			log.Printf("gRPC server error: %v", err)
		}
	}()

	// Return a shutdown function.
	stopFunc = func() {
		grpcServer.GracefulStop()
		database.Close()
		lis.Close()
	}
	return stopFunc, nil
}
