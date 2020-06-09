package main

import (
	"minitube/http"
	"minitube/store"
	"minitube/utils"
)

var log = utils.Sugar

func main() {

	defer log.Sync()
	defer store.CloseAll()

	http.Router.Run(":80")
}