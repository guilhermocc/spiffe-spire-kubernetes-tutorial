package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	"github.com/spiffe/go-spiffe/v2/spiffegrpc/grpccredentials"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/examples/helloworld/helloworld"

	"greeter/cmd"
)

func main() {
	log.Println("Starting up...")
	ctx := context.Background()

	listenAddr := os.Getenv("LISTEN_ADDR")
	wlAPIAddr := os.Getenv("SPIFFE_ENDPOINT_SOCKET")

	var serverOpts []grpc.ServerOption

	if wlAPIAddr != "" {
		log.Println("SPIFFE config enabled, setting up mTLS")

		x509Source := cmd.ConnectToWorkloadAPI(ctx, wlAPIAddr)
		defer x509Source.Close()
		serverOpts = append(serverOpts,
			grpc.Creds(grpccredentials.MTLSServerCredentials(
				x509Source, // SVID source
				x509Source, // Bundle source
				tlsconfig.AuthorizeAny())),
		)
	} else {
		log.Println("SPIFFE config disabled")
		serverOpts = append(serverOpts, grpc.Creds(insecure.NewCredentials()))
	}

	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatal(err)
	}

	server := grpc.NewServer(serverOpts...)
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
	clientID := "SOME-CLIENT-ID"
	if peerID, ok := grpccredentials.PeerIDFromContext(ctx); ok {
		clientID = peerID.String()
	}

	log.Printf("%s said hello %q", clientID, req.Name)

	return &helloworld.HelloReply{
		Message: fmt.Sprintf("Hello, %s!", clientID),
	}, nil
}
