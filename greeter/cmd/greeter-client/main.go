package main

import (
	"context"
	"encoding/pem"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spiffe/go-spiffe/v2/spiffegrpc/grpccredentials"
	"github.com/spiffe/go-spiffe/v2/spiffeid"
	"github.com/spiffe/go-spiffe/v2/spiffetls/tlsconfig"
	"github.com/spiffe/go-spiffe/v2/workloadapi"
	"google.golang.org/grpc"
	"google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/peer"
)

func main() {
	var addr string
	flag.StringVar(&addr, "addr", "", "host:port of the server")
	flag.Parse()

	var authorizedSpiffeIDs string
	flag.StringVar(&authorizedSpiffeIDs, "authorized-spiffe-ids", "", "authorized spiffe IDs separated by comma")
	flag.Parse()
	// spiffe IDs separated by comma

	if authorizedSpiffeIDs == "" {
		authorizedSpiffeIDs = os.Getenv("AUTHORIZED_SPIFFE_IDS")
	}

	var authorizedSpiffeIDsSlice []spiffeid.ID
	for _, sid := range strings.Split(authorizedSpiffeIDs, ",") {
		authSpiffeID, err := spiffeid.FromString(sid)
		if err != nil {
			log.Fatal("Invalid SPIFFE ID %q: %v", sid, err)
		}
		authorizedSpiffeIDsSlice = append(authorizedSpiffeIDsSlice, authSpiffeID)
	}

	if addr == "" {
		addr = os.Getenv("GREETER_SERVER_ADDR")
	}
	if addr == "" {
		addr = "localhost:8080"
	}
	log.Println("Starting up...")
	log.Println("Server Address:", addr)

	ctx := context.Background()

	wlAPIAddrenv := os.Getenv("SPIFFE_ENDPOINT_SOCKET")
	log.Printf("Connecting to Workload API at %q...", wlAPIAddrenv)

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
	bundle, err := source.GetX509BundleForTrustDomain(svid.ID.TrustDomain())
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("SVID: %q", svid.ID)
	log.Printf("Bundles:")
	for _, bundleCert := range bundle.X509Authorities() {
		// 	print in pem format
		pemObject := pem.EncodeToMemory(&pem.Block{
			Type:  "CERTIFICATE",
			Bytes: bundleCert.Raw,
		})

		fmt.Printf("Bundle cert: %s", pemObject)
	}

	creds := grpccredentials.MTLSClientCredentials(source, source, tlsconfig.AuthorizeOneOf(authorizedSpiffeIDsSlice...))

	client, err := grpc.DialContext(ctx, addr, grpc.WithTransportCredentials(creds))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	greeterClient := helloworld.NewGreeterClient(client)

	const interval = time.Second * 10
	log.Printf("Issuing requests every %s...", interval)
	for {
		issueRequest(ctx, greeterClient)
		time.Sleep(interval)
	}
}

func issueRequest(ctx context.Context, c helloworld.GreeterClient) {
	p := new(peer.Peer)
	resp, err := c.SayHello(ctx, &helloworld.HelloRequest{
		Name: "Joe",
	}, grpc.Peer(p))
	if err != nil {
		log.Printf("Failed to say hello: %v", err)
		return
	}

	// ///////////////////////////////////////////////////////////////////////
	// TODO: Obtain the server SPIFFE ID
	// ///////////////////////////////////////////////////////////////////////
	serverID := "SOME-SERVER"
	if peerID, ok := grpccredentials.PeerIDFromPeer(p); ok {
		serverID = peerID.String()
	}

	log.Printf("%s said %q", serverID, resp.Message)
}
