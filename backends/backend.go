package backends

import "github.com/donovanhide/eventsource"

// For callbacks when an invoice gets settled
type PublishInvoiceSettled func(invoice string, eventSrv *eventsource.Server)

type Backend interface {
	Connect() error

	// Amount in satoshi and expiry in seconds
	GetInvoice(description string, amount int64, expiry int64) (invoice string, err error)

	SubscribeInvoices(callback PublishInvoiceSettled, eventSrv *eventsource.Server) error
}
