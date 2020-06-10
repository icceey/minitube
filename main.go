package main

import (
	"minitube/api"
	"minitube/store"
	"minitube/utils"
)

var log = utils.Sugar

func main() {

	defer log.Sync()
	defer store.CloseAll()

	api.Router.Run(":80")
}
