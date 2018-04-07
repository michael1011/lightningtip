package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/donovanhide/eventsource"
	"github.com/michael1011/lightningtip/database"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const eventChannel = "invoiceSettled"

const couldNotParseError = "Could not parse values from request"

var eventSrv *eventsource.Server

var pendingInvoices []PendingInvoice

type PendingInvoice struct {
	Invoice     string
	Amount      int64
	Message     string
	PaymentHash []byte
	Hash        string
	Expiry      time.Time
}

// To use the pendingInvoice type as event for the EventSource stream
func (pending PendingInvoice) Id() string    { return "" }
func (pending PendingInvoice) Event() string { return "" }
func (pending PendingInvoice) Data() string  { return pending.Hash }

type invoiceRequest struct {
	Amount  int64
	Message string
}

type invoiceResponse struct {
	Invoice string
	Expiry  int64
}

type invoiceSettledRequest struct {
	InvoiceHash string
}

type invoiceSettledResponse struct {
	Settled bool
}

type errorResponse struct {
	Error string
}

// TODO: add option to show URI of Lightning node
func main() {
	initLog()

	initConfig()

	dbErr := database.InitDatabase(cfg.DatabaseFile)

	if dbErr != nil {
		log.Error("Failed to initialize database: " + fmt.Sprint(dbErr))

		os.Exit(1)

	} else {
		log.Debug("Opened SQLite database: " + cfg.DatabaseFile)
	}

	err := backend.Connect()

	if err == nil {
		log.Info("Starting EventSource stream")

		eventSrv = eventsource.NewServer()

		http.Handle("/", handleHeaders(notFoundHandler))
		http.Handle("/getinvoice", handleHeaders(getInvoiceHandler))
		http.Handle("/eventsource", handleHeaders(eventSrv.Handler(eventChannel)))

		// Alternative for browsers which don't support EventSource (Internet Explorer and Edge)
		http.Handle("/invoicesettled", handleHeaders(invoiceSettledHandler))

		log.Debug("Starting ticker to clear expired invoices")

		// A bit longer than the expiry time to make sure the invoice doesn't show as settled if it isn't (affects just invoiceSettledHandler)
		expiryInterval := time.Duration(cfg.TipExpiry + 10)
		expiryTicker := time.Tick(expiryInterval * time.Second)

		go func() {
			for {
				select {
				case <-expiryTicker:
					now := time.Now()

					for index := len(pendingInvoices) - 1; index >= 0; index-- {
						invoice := pendingInvoices[index]

						if now.Sub(invoice.Expiry) > 0 {
							log.Debug("Invoice expired: " + invoice.Invoice)

							pendingInvoices = append(pendingInvoices[:index], pendingInvoices[index+1:]...)
						}

					}

				}

			}

		}()

		log.Info("Subscribing to invoices")

		go func() {
			subscribeToInvoices()
		}()

		if cfg.KeepAliveInterval > 0 {
			log.Debug("Starting ticker to send keepalive requests")

			interval := time.Duration(cfg.KeepAliveInterval)
			keepAliveTicker := time.Tick(interval * time.Second)

			go func() {
				for {
					select {
					case <-keepAliveTicker:
						backend.KeepAliveRequest()
					}
				}
			}()

		}

		log.Info("Starting HTTP server")

		go func() {
			var err error

			if cfg.TLSCertFile != "" && cfg.TLSKeyFile != "" {
				err = http.ListenAndServeTLS(cfg.RESTHost, cfg.TLSCertFile, cfg.TLSKeyFile, nil)

			} else {
				err = http.ListenAndServe(cfg.RESTHost, nil)
			}

			if err != nil {
				log.Errorf("Failed to start HTTP server: " + fmt.Sprint(err))

				os.Exit(1)
			}

		}()

		select {}

	}

}

func subscribeToInvoices() {
	err := backend.SubscribeInvoices(publishInvoiceSettled, rescanPendingInvoices)

	log.Error("Failed to subscribe to invoices: " + fmt.Sprint(err))

	if err != nil {
		if cfg.ReconnectInterval != 0 {
			time.Sleep(time.Duration(cfg.ReconnectInterval) * time.Second)

			log.Info("Trying to reconnect to LND")

			subscribeToInvoices()

		} else {
			os.Exit(1)
		}

	}

}

func rescanPendingInvoices() {
	log.Info("Connected to LND")

	if len(pendingInvoices) > 0 {
		log.Debug("Rescanning pending invoices")

		for _, invoice := range pendingInvoices {
			settled, err := backend.InvoiceSettled(invoice.PaymentHash)

			if err == nil {
				if settled {
					publishInvoiceSettled(invoice.Invoice)
				}

			} else {
				log.Warning("Failed to check if invoice settled \"" + invoice.Invoice + "\": " + fmt.Sprint(err))
			}

		}

	}

}

func publishInvoiceSettled(invoice string) {
	for index, pending := range pendingInvoices {
		if pending.Invoice == invoice {
			log.Info("Invoice settled: " + invoice)

			eventSrv.Publish([]string{eventChannel}, pending)

			database.AddSettledInvoice(pending.Amount, pending.Message)

			pendingInvoices = append(pendingInvoices[:index], pendingInvoices[index+1:]...)

			break
		}

	}

}

func invoiceSettledHandler(writer http.ResponseWriter, request *http.Request) {
	errorMessage := couldNotParseError

	if request.Method == http.MethodPost {
		var body invoiceSettledRequest

		data, _ := ioutil.ReadAll(request.Body)

		err := json.Unmarshal(data, &body)

		if err == nil {
			if body.InvoiceHash != "" {
				settled := true

				for _, pending := range pendingInvoices {
					if pending.Hash == body.InvoiceHash {
						settled = false

						break
					}

				}

				writer.Write(marshalJson(invoiceSettledResponse{
					Settled: settled,
				}))

				return

			}

		}

	}

	log.Error(errorMessage)

	writeError(writer, errorMessage)
}

func getInvoiceHandler(writer http.ResponseWriter, request *http.Request) {
	errorMessage := couldNotParseError

	if request.Method == http.MethodPost {
		var body invoiceRequest

		data, _ := ioutil.ReadAll(request.Body)

		err := json.Unmarshal(data, &body)

		if err == nil {
			if body.Amount != 0 {
				invoice, paymentHash, err := backend.GetInvoice(body.Message, body.Amount, cfg.TipExpiry)

				if err == nil {
					logMessage := "Created invoice with amount of " + strconv.FormatInt(body.Amount, 10) + " satoshis"

					if body.Message != "" {
						// Deletes new lines at the end of the messages
						body.Message = strings.TrimSuffix(body.Message, "\n")

						logMessage += " with message \"" + body.Message + "\""
					}

					sha := sha256.New()
					sha.Write([]byte(invoice))

					hash := hex.EncodeToString(sha.Sum(nil))

					expiryDuration := time.Duration(cfg.TipExpiry) * time.Second

					log.Info(logMessage)

					pendingInvoices = append(pendingInvoices, PendingInvoice{
						Invoice:     invoice,
						Amount:      body.Amount,
						Message:     body.Message,
						PaymentHash: paymentHash,
						Hash:        hash,
						Expiry:      time.Now().Add(expiryDuration),
					})

					writer.Write(marshalJson(invoiceResponse{
						Invoice: invoice,
						Expiry:  cfg.TipExpiry,
					}))

					return

				} else {
					errorMessage = "Failed to create invoice"

					// This is way too hacky
					// Maybe a cast to the gRPC error and get its error message directly
					if fmt.Sprint(err)[:47] == "rpc error: code = Unknown desc = memo too large" {
						errorMessage += ": message too long"
					}

				}

			}

		}

	}

	log.Error(errorMessage)

	writeError(writer, errorMessage)
}

func notFoundHandler(writer http.ResponseWriter, request *http.Request) {
	if request.RequestURI == "/" {
		writeError(writer, "This is an API to connect LND and your website. You should not open this in your browser")

	} else {
		writeError(writer, "Not found")
	}
}

func handleHeaders(handler func(w http.ResponseWriter, r *http.Request)) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if cfg.AccessDomain != "" {
			writer.Header().Add("Access-Control-Allow-Origin", cfg.AccessDomain)
		}

		handler(writer, request)
	})
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
