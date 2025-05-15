package main

import (
	app "github.com/genefriendway/human-network-iam/cmd/app"
	"github.com/genefriendway/human-network-iam/conf"
	_ "github.com/genefriendway/human-network-iam/docs"
)

func main() {
	config := conf.GetConfiguration()
	app.RunApp(config)
}
