package backends

type Backend interface {
	Connect() error

	// Value in satoshis and expiry in seconds
	GetInvoice(description string, value int64, expiry int64) (invoice string, err error)
}
