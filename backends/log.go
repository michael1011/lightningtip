package backends

import "github.com/op/go-logging"

var log logging.Logger

func UseLogger(logger logging.Logger) {
	log = logger
}
