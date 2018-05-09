package backends

// For callbacks when an invoice gets settled
type PublishInvoiceSettled func(invoice string)

// For callbacks when reconnecting
type RescanPendingInvoices func()

type Backend interface {
	Connect() error

	// Amount in satoshi and expiry in seconds
	GetInvoice(description string, amount int64, expiry int64) (invoice string, rHash string, err error)

	InvoiceSettled(rHash string) (settled bool, err error)

	SubscribeInvoices(publish PublishInvoiceSettled, rescan RescanPendingInvoices) error

	KeepAliveRequest() error
}
