package main

import (
	"github.com/vartanbeno/go-reddit/v2/reddit"
)

func main() {
	OpenDB()

	var err error
	client, err = reddit.NewReadonlyClient()
	if err != nil {
		panic(err)
	}

	StartUpdater("battlemaps")
	Downloader()
}
