package main

import (
	"github.com/alexraskin/goShareX/server"

	"github.com/syumai/workers"
	"github.com/syumai/workers/cloudflare"
)

// cloudflare binding
const bucketName = "IMAGE_BUCKET"

var authKey = cloudflare.Getenv("SHAREX_AUTH_KEY")

func main() {
	workers.Serve(&server.Server{
		BucketName: bucketName,
		AuthKey:    authKey,
	})
}
