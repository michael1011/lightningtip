package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

type invoiceResponse struct {
	Invoice string
	Expiry  int64
}

type invoiceRequest struct {
	Amount  int64
	Message string
}

type errorResponse struct {
	Error string
}

// TODO: add option to show URI of Lightning node
func main() {
	initLog()

	initConfig()

	err := backend.Connect()

	if err == nil {
		http.HandleFunc("/", notFoundHandler)
		http.HandleFunc("/getinvoice", getInvoiceHandler)

		log.Info("Subscribing to invoices")

		go func() {
			// TODO: let clients listen if their invoice was paid (eventsource)
			err = cfg.LND.SubscribeInvoices()

			if err != nil {
				log.Error("Failed to subscribe to invoices: " + fmt.Sprint(err))

				os.Exit(1)
			}

		}()

		log.Info("Starting HTTP server")

		go func() {
			err = http.ListenAndServe(cfg.RESTHost, nil)

			if err != nil {
				log.Error("Failed to start HTTP server: " + fmt.Sprint(err))

				os.Exit(1)
			}

		}()

		select {}

	}

}

func getInvoiceHandler(writer http.ResponseWriter, request *http.Request) {
	errorMessage := "Could not parse values from request"

	if request.Method == http.MethodPost {
		var body invoiceRequest

		data, _ := ioutil.ReadAll(request.Body)

		err := json.Unmarshal(data, &body)

		if err == nil {
			if body.Amount != 0 {
				invoice, err := backend.GetInvoice(body.Message, body.Amount, cfg.TipExpiry)

				if err == nil {
					logMessage := "Created invoice with amount of " + strconv.FormatInt(body.Amount, 10) + " satoshis"

					if body.Message != "" {
						logMessage += " with message \"" + body.Message + "\""
					}

					log.Info(logMessage)

					writer.Write(marshalJson(invoiceResponse{
						Invoice: invoice,
						Expiry:  cfg.TipExpiry,
					}))

					return

				} else {
					errorMessage = "Failed to create invoice"
				}

			}

		}

	}

	log.Error(errorMessage)

	writeError(writer, errorMessage)
}

func notFoundHandler(writer http.ResponseWriter, request *http.Request) {
	writeError(writer, "Not found")
}

func writeError(writer http.ResponseWriter, message string) {
	writer.WriteHeader(http.StatusBadRequest)

	writer.Write(marshalJson(errorResponse{
		Error: message,
	}))
}

func marshalJson(data interface{}) []byte {
	response, _ := json.MarshalIndent(data, "", "    ")

	return response
}
