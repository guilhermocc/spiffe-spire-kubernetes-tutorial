package cmd

import (
	"context"
	"log"

	"github.com/spiffe/go-spiffe/v2/workloadapi"
)

func ConnectToWorkloadAPI(ctx context.Context, wlAPIAddr string) *workloadapi.X509Source {
	log.Printf("Connecting to Workload API at %q...", wlAPIAddr)

	source, err := workloadapi.NewX509Source(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Connected to Workload API at %q", wlAPIAddr)

	svid, err := source.GetX509SVID()
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("SPIFFE ID: %q", svid.ID)

	return source
}
