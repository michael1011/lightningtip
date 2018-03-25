package backends

type Backend interface {
	Connect() error

	// Amount in satoshis and expiry in seconds
	GetInvoice(description string, amount int64, expiry int64) (invoice string, err error)
}
