package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/spiffe/go-spiffe/v2/spiffegrpc/grpccredentials"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/examples/helloworld/helloworld"
)

func main() {
	var addr string
	flag.StringVar(&addr, "addr", "localhost:8080", "host:port of the server")
	flag.Parse()

	log.Println("Starting up...")

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	wlAPIAddrenv := os.Getenv("SPIFFE_ENDPOINT_SOCKET")
	log.Printf("Connecting to Workload API at %q...", wlAPIAddrenv)

	// SPIFFE_ENDPOINT_SOCKET environment variable will be used.
	source, err := workloadapi.NewX509Source(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer source.Close()

	log.Printf("Connected to Workload API at %q", wlAPIAddrenv)

	svid, err := source.GetX509SVID()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("SPIFFE ID: %q", svid.ID)

	creds := grpccredentials.MTLSServerCredentials(source, source, tlsconfig.AuthorizeAny())

	server := grpc.NewServer(grpc.Creds(creds))
	helloworld.RegisterGreeterServer(server, greeter{})

	log.Println("Serving on", listener.Addr())
	if err := server.Serve(listener); err != nil {
		log.Fatal(err)
	}
}

type greeter struct {
	helloworld.UnimplementedGreeterServer
}

func (greeter) SayHello(ctx context.Context, req *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	clientID := "SOME-CLIENT"
	if peerID, ok := grpccredentials.PeerIDFromContext(ctx); ok {
		clientID = peerID.String()
	}

	log.Printf("%s has requested that I say say hello to %q...", clientID, req.Name)
	return &helloworld.HelloReply{
		Message: fmt.Sprintf("On behalf of %s, hello %s!", clientID, req.Name),
	}, nil
}
