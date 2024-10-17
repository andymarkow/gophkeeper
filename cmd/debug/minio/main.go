package main

import (
	"context"
	"log"
	"os"

	"github.com/andymarkow/gophkeeper/internal/storage/objrepo"
)

func main() {
	client, err := objrepo.NewMinioClient("localhost:9000", "mybucket", &objrepo.MinioClientOpts{
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin",
		UseSSL:          false,
	})
	if err != nil {
		log.Fatalf("objrepo.NewMinioClient: %v", err)
	}

	ctx := context.Background()

	if err := client.InitBucket(ctx); err != nil {
		log.Fatalf("client.InitBucket: %v", err)
	}

	f, err := os.Open("testdata/simple/data.json")
	if err != nil {
		log.Fatalf("os.Open: %v", err)
	}
	defer f.Close()

	info, err := client.PutObject(ctx, "myuser/data.json", -1, f)
	if err != nil {
		log.Fatalf("client.PutObject: %v", err)
	}

	log.Printf("Object info: %+v", info)
}
