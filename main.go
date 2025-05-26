package main

import (
	"insider-case/app"
	"insider-case/config"
)

func main() {
	config, err := config.LoadConfig()

	if err != nil {
		panic(err)
	}

	appInstance := &app.App{}
	appInstance.Initialize(config)
	appInstance.Run()
}
