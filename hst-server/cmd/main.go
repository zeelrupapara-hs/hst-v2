package main

import (
	"hstserver/app"
)

// @title						HS Trader
// @version					1.0
// @description				HS Trader APIs doc
// @termsOfService				https://hybridsolutions.com
// @contact.name				API Support
// @contact.email				support@hybridsolutions.com
// @license.url				https://hybridsolutions.com
// @license.name				HS Trader
// @BasePath					/
// @Schemes					http https
// @securityDefinitions.basic	BasicAuth
// @name						Authorization
func main() {

	app.Start()
}
