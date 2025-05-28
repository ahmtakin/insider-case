package main

import (
	"fmt"
	"insider-case/app"
	"insider-case/config"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("Application panicked: %v\n", r)
		}
	}()
	fmt.Println("Starting Insider Case Application...")
	config, err := config.LoadConfig()

	if err != nil {
		panic(err)
	}

	appInstance := &app.App{}
	appInstance.Initialize(config)
	appInstance.Run()
}
