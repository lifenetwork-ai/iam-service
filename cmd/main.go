package main

import (
	app "github.com/genefriendway/human-network-auth/cmd/app"
	"github.com/genefriendway/human-network-auth/conf"
	_ "github.com/genefriendway/human-network-auth/docs"
)

func main() {
	config := conf.GetConfiguration()
	app.RunApp(config)
}
