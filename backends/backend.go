package backends

// PublishInvoiceSettled is a callback for a settled invoice
type PublishInvoiceSettled func(invoice string)

// RescanPendingInvoices is a callbacks when reconnecting
type RescanPendingInvoices func()

// Backend is an interface that would allow for different implementations of Lightning to be used as backend
type Backend interface {
	Connect() error

	// The amount is denominated in satoshis and the expiry in seconds
	GetInvoice(description string, amount int64, expiry int64) (invoice string, rHash string, picture string, err error)

	InvoiceSettled(rHash string) (settled bool, err error)

	SubscribeInvoices(publish PublishInvoiceSettled, rescan RescanPendingInvoices) error

	KeepAliveRequest() error
}
