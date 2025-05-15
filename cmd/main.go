package main

import (
	app "github.com/lifenetwork-ai/iam-service/cmd/app"
	"github.com/lifenetwork-ai/iam-service/conf"
	_ "github.com/lifenetwork-ai/iam-service/docs"
)

func main() {
	config := conf.GetConfiguration()
	app.RunApp(config)
}
