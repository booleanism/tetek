package docs

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
)

type ServiceName int

const (
	Feeds ServiceName = iota
)

var serviceMap = map[ServiceName]string{
	Feeds: "/feeds.yaml",
}

func OapiDocs(r fiber.Router, sn ServiceName, endp string) (func() string, string) {
	path := "docs/" + endp + serviceMap[sn]
	return func() string {
			f, err := os.ReadFile(path)
			if err != nil {
				log.Fatal(err.Error())
			}
			return string(f)
		}, `<!doctype html>
	<html lang="en">
	
	 <head>
	   <meta charset="utf-8">
	   <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
	   <title>Elements in HTML</title>
	   <!-- Embed elements Elements via Web Component -->
	   <script src="https://unpkg.com/@stoplight/elements/web-components.min.js"></script>
	   <link rel="stylesheet" href="https://unpkg.com/@stoplight/elements/styles.min.css">
	 </head>
	 <body>
	
	   <elements-api
	     apiDescriptionUrl="` + endp + "/openapi.yaml" + `"
	     router="hash"
	     layout="sidebar"
	   />
	
	 </body>
	</html>`
}
