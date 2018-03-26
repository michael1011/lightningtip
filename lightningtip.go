package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/donovanhide/eventsource"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"
)

const eventChannel = "invoiceSettled"

var eventSrv *eventsource.Server

var pendingInvoices []PendingInvoice

type PendingInvoice struct {
	Invoice string
	Hash    string
	Expiry  time.Time
}

// To use the pendingInvoice type as event for the EventSource stream
func (pending PendingInvoice) Id() string    { return "" }
func (pending PendingInvoice) Event() string { return "" }
func (pending PendingInvoice) Data() string  { return pending.Hash }

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
		log.Info("Starting EventSource stream")

		eventSrv = eventsource.NewServer()

		http.HandleFunc("/", notFoundHandler)
		http.HandleFunc("/getinvoice", getInvoiceHandler)
		http.HandleFunc("/eventsource", eventSrv.Handler(eventChannel))

		log.Info("Subscribing to invoices")

		go func() {
			err = cfg.LND.SubscribeInvoices(publishInvoiceSettled, eventSrv)

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

// Callbacks when an invoice gets settled
func publishInvoiceSettled(invoice string, eventSrv *eventsource.Server) {
	for index, pending := range pendingInvoices {
		if pending.Invoice == invoice {
			log.Info("Invoice settled: " + invoice)

			eventSrv.Publish([]string{eventChannel}, pending)

			pendingInvoices = append(pendingInvoices[:index], pendingInvoices[index+1:]...)

			break
		}

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

					sha := sha256.New()
					sha.Write([]byte(invoice))

					hash := hex.EncodeToString(sha.Sum(nil))

					expiryDuration := time.Duration(cfg.TipExpiry) * time.Second

					log.Info(logMessage)

					// TODO: check every minute or so if expired
					pendingInvoices = append(pendingInvoices, PendingInvoice{
						Invoice: invoice,
						Hash:    hash,
						Expiry:  time.Now().Add(expiryDuration),
					})

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
