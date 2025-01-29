package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spiffe/go-spiffe/v2/spiffegrpc/grpccredentials"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"google.golang.org/grpc"
	"google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/peer"

	"greeter/cmd"
)

func main() {
	log.Println("Starting up...")
	ctx := context.Background()

	addr := os.Getenv("GREETER_SERVER_ADDR")
	authorizedSpiffeIDs := os.Getenv("AUTHORIZED_SPIFFE_IDS")
	wlAPIAddr := os.Getenv("SPIFFE_ENDPOINT_SOCKET")

	var dialOpts []grpc.DialOption

	if wlAPIAddr != "" {
		log.Println("SPIFFE config enabled, setting up mTLS")

		var authorizedSpiffeIDsSlice []spiffeid.ID
		for _, sid := range strings.Split(authorizedSpiffeIDs, ",") {
			authSpiffeID, err := spiffeid.FromString(sid)
			if err != nil {
				log.Fatal("Invalid SPIFFE ID %q: %v", sid, err)
			}
			authorizedSpiffeIDsSlice = append(authorizedSpiffeIDsSlice, authSpiffeID)
		}

		x509Source := cmd.ConnectToWorkloadAPI(ctx, wlAPIAddr)
		defer x509Source.Close()
		dialOpts = append(dialOpts, grpc.WithTransportCredentials(
			grpccredentials.MTLSClientCredentials(
				x509Source, // SVID source
				x509Source, // Bundle source
				tlsconfig.AuthorizeOneOf(authorizedSpiffeIDsSlice...))))
	} else {
		log.Println("SPIFFE config disabled")
		dialOpts = append(dialOpts, grpc.WithInsecure())
	}

	log.Println("Server Address:", addr)

	client, err := grpc.DialContext(ctx, addr, dialOpts...)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	greeterClient := helloworld.NewGreeterClient(client)

	const interval = 1 * time.Second
	log.Printf("Making SayHello requests every %s...", interval)
	for {
		issueRequest(ctx, greeterClient)
		time.Sleep(interval)
	}
}

func issueRequest(ctx context.Context, c helloworld.GreeterClient) {
	p := new(peer.Peer)
	resp, err := c.SayHello(ctx, &helloworld.HelloRequest{
		Name: "Server",
	}, grpc.Peer(p))
	if err != nil {
		log.Printf("Failed to say hello: %v", err)
		return
	}

	serverID := "SOME-SERVER-ID"
	if peerID, ok := grpccredentials.PeerIDFromPeer(p); ok {
		serverID = peerID.String()
	}

	log.Printf("%s said %q", serverID, resp.Message)
}
