package main

import "fmt"

func main() {
	initLog()

	initConfig()

	err := backend.Connect()

	if err == nil {
		log.Info("Successfully connected to " + backendName)

		invoice, err := backend.GetInvoice("Just a test", 1, 3600)

		log.Info("Got invoice " + invoice)
		log.Info(fmt.Sprint(err))
	}

}
