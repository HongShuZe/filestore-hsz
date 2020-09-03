package main

import "filestore-hsz/service/apiwg/route"

func main() {
	r := route.Router()
	r.Run(":8080")
}
