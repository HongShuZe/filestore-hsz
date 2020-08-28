package main

import (
	"filestore-hsz/route"
	cfg "filestore-hsz/config"
)

func main()  {
	router := route.Router()
	router.Run(cfg.UploadServiceHost)
}