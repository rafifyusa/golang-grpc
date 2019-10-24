package main

import (
	"context"
	"fmt"
	"golangbootcamp/gRPC/greet/greetpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"time"
)

func main() {
	fmt.Println("Hello I'm client")
	cc, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Couldn connect: %v", err)
	}
	defer cc.Close()

	c := greetpb.NewGreetServiceClient(cc)
	//fmt.Printf("Created client: %f", c)

	//doUnary(c)
	//doServerStreaming(c)
	//doClientStreaming(c)
	//doBidirectionalStreaming(c)
	doUnaryWithDeadline(c, 1 * time.Second)//should complete
	doUnaryWithDeadline(c, 5 * time.Second)//should timeout
}

func doBidirectionalStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Start doing bidirectional streaming RPC")

	//we create a stream by invoking client
	stream, err := c.GreetEveryone(context.Background())
	if err != nil {
		log.Fatalf("Error while creating stream: %v", err)
		return
	}

	requests := []*greetpb.GreetEveryoneRequest {
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName:"Stephane",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName:"Edward",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName:"Alphonse",
			},
		},
		&greetpb.GreetEveryoneRequest{
			Greeting: &greetpb.Greeting{
				FirstName:"Mustang",
			},
		},
	}

	waitc := make(chan struct{})
	//we send a bunch of messages to the client (go routine)
	go func() {
		//function to send a bunch of messages
		for _, req:= range requests {
			fmt.Printf("Sending message: %v\n", req)
			stream.Send(req)
			time.Sleep(1000 * time.Millisecond)
		}
		stream.CloseSend()
	}()
	//we receive a bunch of messages from the client (go routine)
	go func() {
		//function to receive a bunch of messages
		for {
			res, err:= stream.Recv()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Fatalf("Error while receiving: %v", err)
			}
			time.Sleep(2000 * time.Millisecond)
			fmt.Printf("Receive: %v\n", res.GetResult())
		}
		close(waitc)
	}()
	//block until everything is done
	<-waitc
}

func doUnary(c greetpb.GreetServiceClient) {
	fmt.Println("Start doing unary RPC")
	req := &greetpb.GreetRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Vaan",
			LastName: "Hohenheim",
		},
	}
	res, err := c.Greet(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling Greet RPC: %v", err)
	}
	log.Printf("Response from Greet: %v", res.Result)
}

func doServerStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Start doing server streaming RPC")

	req := &greetpb.GreetManyTimesRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Vaan",
			LastName: "Hohenheim",
		},
	}
	resStream, err := c.GreetManyTimes(context.Background(), req)
	if err != nil {
		log.Fatalf("Error while calling GreetManyTimes RPC: %v", err)
	}
	for {
		msg, err := resStream.Recv()
		if err == io.EOF {
			// reached end of stream
			break
		}
		if err != nil {
			log.Fatalf("Error while reading stream : %v", err)
		}
		log.Printf("Response from Greet: %v", msg.GetResult())
	}
}

func doClientStreaming(c greetpb.GreetServiceClient) {
	fmt.Println("Start doing client streaming RPC")

	requests := []*greetpb.LongGreetRequest {
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName:"Stephane",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName:"Edward",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName:"Alphonse",
			},
		},
		&greetpb.LongGreetRequest{
			Greeting: &greetpb.Greeting{
				FirstName:"Mustang",
			},
		},
	}

	stream, err := c.LongGreet(context.Background())
	if err != nil {
		log.Fatalf("Error while calling LongGreet RPC: %v", err)
	}

	// Iterate over slice and send each message individually
	for _, req := range requests {
		fmt.Printf("Sending req: %v\n", req)
		stream.Send(req)
		time.Sleep(1000 * time.Millisecond)
	}
	res, err:= stream.CloseAndRecv()
	if err!= nil {
		log.Fatalf("Error while receiving response from LongGreet : %v", err)
	}
	fmt.Printf("LongGreet Response: %v\n", res)
}

func doUnaryWithDeadline(c greetpb.GreetServiceClient, timeout time.Duration) {
	fmt.Println("Start doing UnaryWithDeadline RPC")
	req := &greetpb.GreetWithDeadlineRequest{
		Greeting: &greetpb.Greeting{
			FirstName: "Vaan",
			LastName: "Hohenheim",
		},
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	res, err := c.GreetWithDeadline(ctx, req)
	if err != nil {

		statusErr, ok := status.FromError(err)
		if ok {
			if statusErr.Code() == codes.DeadlineExceeded {
				fmt.Println("Timeout was hit! Deadline was exceeded")
			} else {
				fmt.Printf("Unexoected error: %v\n", err)
			}
		} else {
			log.Fatalf("Error while calling GreetWithDeadline RPC: %v", err)
		}
		return
	}
	log.Printf("Response from GreetWithDeadline: %v", res.Result)
}