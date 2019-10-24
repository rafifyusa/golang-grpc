package main

import (
	"context"
	"fmt"
	"golangbootcamp/gRPC/blog/blogpb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"os/signal"
	"time"
)

var collection *mongo.Collection

type server struct {}

type blogItem struct {
	ID objectid.ObjectID `bson:"_id,omitempty"`
	AuthorID string `bson:"author_id"`
	Content string `bson"content"`
	Title string `bson:"title"`

}
func main() {
	//if we crash the go code, we get the file name and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	fmt.Println("Hello Blog Server..!")

	//initiating mongo client
	fmt.Println("Connecting to MongoDB")
	client, err := mongo.NewClient(options.Client().ApplyURI("mongodb://@localhost:27017"))
	if err != nil { log.Fatal(err) }
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil { log.Fatal(err) }

	collection = client.Database("mydb").Collection("blog")

	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	opts := []grpc.ServerOption{}
	s:= grpc.NewServer(opts...)
	blogpb.RegisterBlogServiceServer(s , &server{})

	go func() {
		fmt.Println("Starting server...")
		if err:= s.Serve(lis); err!= nil {
			log.Fatalf("Failed to serve: %v", err)
		}
	}()

	//wait for ctrl+c to exit
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)

	//block until a signal is received
	<-ch
	fmt.Println("Stopping the server")
	s.Stop()
	fmt.Println("Closing the listener")
	lis.Close()
	fmt.Println("Closing MongoDB connection")
	client.Disconnect(context.Background())
	fmt.Println("End of Program")
	fmt.Println("===============")
}
