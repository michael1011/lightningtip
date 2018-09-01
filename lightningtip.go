package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/donovanhide/eventsource"
	"github.com/michael1011/lightningtip/database"
)

// PendingInvoice is for keeping alist of unpaid invoices
type PendingInvoice struct {
	Invoice string
	Amount  int64
	Message string
	RHash   string
	Expiry  time.Time
}

const eventChannel = "invoiceSettled"

const couldNotParseError = "Could not parse values from request"

var eventSrv *eventsource.Server

var pendingInvoices []PendingInvoice

// To use the pendingInvoice type as event for the EventSource stream
func (pending PendingInvoice) Id() string    { return "" }
func (pending PendingInvoice) Event() string { return "" }
func (pending PendingInvoice) Data() string  { return pending.RHash }

type invoiceRequest struct {
	Amount  int64
	Message string
}

type invoiceResponse struct {
	Invoice string
	RHash   string
	Expiry  int64
}

type invoiceSettledRequest struct {
	RHash string
}

type invoiceSettledResponse struct {
	Settled bool
}

type errorResponse struct {
	Error string
}

// TODO: add version flag
// TODO: don't start when "--help" flag is provided
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

		defer eventSrv.Close()

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
	log.Info("Subscribing to invoices")

	err := backend.SubscribeInvoices(publishInvoiceSettled, rescanPendingInvoices)

	log.Error("Failed to subscribe to invoices: " + fmt.Sprint(err))

	if err != nil {
		if cfg.ReconnectInterval != 0 {
			reconnectToBackend()

		} else {
			os.Exit(1)
		}

	}

}

func reconnectToBackend() {
	time.Sleep(time.Duration(cfg.ReconnectInterval) * time.Second)

	log.Info("Trying to reconnect to LND")

	backend = cfg.LND

	err := backend.Connect()

	if err == nil {
		err = backend.KeepAliveRequest()

		// The default macaroon file used by LightningTip "invoice.macaroon" allows only creating and checking status of invoices
		// The keep alive request doesn't have to be successful as long as it can establish a connection to LND
		if err == nil || fmt.Sprint(err) == "rpc error: code = Unknown desc = permission denied" {
			log.Info("Reconnected to LND")

			subscribeToInvoices()
		}

	}

	log.Info("Connection failed")

	log.Debug(fmt.Sprint(err))

	reconnectToBackend()
}

func rescanPendingInvoices() {
	if len(pendingInvoices) > 0 {
		log.Debug("Rescanning pending invoices")

		for _, invoice := range pendingInvoices {
			settled, err := backend.InvoiceSettled(invoice.RHash)

			if err == nil {
				if settled {
					publishInvoiceSettled(invoice.Invoice)
				}

			} else {
				log.Warning("Failed to check if invoice settled: " + fmt.Sprint(err))
			}

		}

	}

}

func publishInvoiceSettled(invoice string) {
	for index, settled := range pendingInvoices {
		if settled.Invoice == invoice {
			log.Info("Invoice settled: " + invoice)

			eventSrv.Publish([]string{eventChannel}, settled)

			database.AddSettledInvoice(settled.Amount, settled.Message)

			if cfg.Mail.Recipient != "" {
				cfg.Mail.SendMail(settled.Amount, settled.Message)
			}

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
			if body.RHash != "" {
				settled := true

				for _, pending := range pendingInvoices {
					if pending.RHash == body.RHash {
						settled = false

						break
					}

				}

				writer.Write(marshalJSON(invoiceSettledResponse{
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

					expiryDuration := time.Duration(cfg.TipExpiry) * time.Second

					log.Info(logMessage)

					pendingInvoices = append(pendingInvoices, PendingInvoice{
						Invoice: invoice,
						Amount:  body.Amount,
						Message: body.Message,
						RHash:   string(paymentHash),
						Expiry:  time.Now().Add(expiryDuration),
					})

					writer.Write(marshalJSON(invoiceResponse{
						Invoice: invoice,
						RHash:   paymentHash,
						Expiry:  cfg.TipExpiry,
					}))

					return
				}

				errorMessage = "Failed to create invoice"

				// This is way too hacky
				// Maybe a cast to the gRPC error and get its error message directly
				if fmt.Sprint(err)[:47] == "rpc error: code = Unknown desc = memo too large" {
					errorMessage += ": message too long"
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

	writer.Write(marshalJSON(errorResponse{
		Error: message,
	}))
}

func marshalJSON(data interface{}) []byte {
	response, _ := json.MarshalIndent(data, "", "    ")

	return response
}
