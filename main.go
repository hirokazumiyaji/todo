package main

import (
	"log"
	"net"
	"os"

	"github.com/hirokazumiyaji/todo/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	port := "9090"
	if v := os.Getenv("PORT"); v != "" {
		port = v
	}
	listen, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	server := grpc.NewServer()
	proto.RegisterTodoServiceServer(server, NewTodoServiceServer())
	reflection.Register(server)
	log.Println("listen to port " + port)
	if err := server.Serve(listen); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
