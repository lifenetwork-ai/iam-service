package main

import (
	app "github.com/lifenetwork-ai/iam-service/cmd/app"
	"github.com/lifenetwork-ai/iam-service/conf"
	_ "github.com/lifenetwork-ai/iam-service/docs"
)

// @title IAM Service API
// @version 1.0
// @description Identity and Access Management Service
// @termsOfService http://swagger.io/terms/
// @contact.name API Support
// @contact.email support@lifenetwork.ai
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @BasePath /
// @securityDefinitions.basic BasicAuth
func main() {
	config := conf.GetConfiguration()
	app.RunApp(config)
}
