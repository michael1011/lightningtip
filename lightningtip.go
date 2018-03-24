package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
)

type getInvoiceRequest struct {
	Value   int64
	Message string
}

type invoiceResponse struct {
	Invoice string
	Error   string
}

type tipValueResponse struct {
	TipValue int64
}

type errorResponse struct {
	Error string
}

func main() {
	initLog()

	initConfig()

	err := backend.Connect()

	if err == nil {
		http.HandleFunc("/", notFoundHandler)
		http.HandleFunc("/getinvoice", getInvoiceHandler)
		http.HandleFunc("/defaulttipvalue", defaultTipValueHandler)

		log.Info("Subscribing to invoices")

		go func() {
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
	var errorMessage string

	tipValue := cfg.DefaultTipValue
	tipMessage := cfg.TipMessage

	if request.Method == http.MethodPost {
		var body getInvoiceRequest

		data, _ := ioutil.ReadAll(request.Body)

		err := json.Unmarshal(data, &body)

		if err == nil {
			if body.Value != 0 {
				tipValue = body.Value
			}

			if body.Message != "" {
				tipMessage = body.Message
			}

		} else {
			errorMessage = "Could not parse values from request"

			log.Warning(errorMessage)
		}

	}

	invoice, err := backend.GetInvoice(tipMessage, tipValue, cfg.TipExpiry)

	if err == nil {
		log.Info("Created invoice with value of " + strconv.FormatInt(tipValue, 10) + " satoshis")

		writer.Write(marshalJson(invoiceResponse{
			Invoice: invoice,
			Error:   errorMessage,
		}))

	} else {
		errorMessage := "Failed to create invoice"

		log.Error(errorMessage + ": " + fmt.Sprint(err))

		writer.Write(marshalJson(errorResponse{
			Error: errorMessage,
		}))
	}

}

func defaultTipValueHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Write(marshalJson(tipValueResponse{
		TipValue: cfg.DefaultTipValue,
	}))
}

func notFoundHandler(writer http.ResponseWriter, request *http.Request) {
	writer.Write(marshalJson(errorResponse{
		Error: "Not found",
	}))
}

func marshalJson(data interface{}) []byte {
	response, _ := json.MarshalIndent(data, "", "    ")

	return response
}
