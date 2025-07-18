package errvalues

import "errors"

var (
	ErrNoSuchRow = errors.New("lack of row with such id")
)
