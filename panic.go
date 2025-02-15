package orz

import (
	"fmt"

	"github.com/go-errors/errors"
)

type PanicErr struct {
	PanicObj interface{}
}

func (e PanicErr) Error() string {
	return fmt.Sprintf("%#v", e.PanicObj)
}

func interfaceToError(i interface{}) error {
	switch out := i.(type) {
	case error:
		return out
	case string:
		return errors.New(out)
	default:
		return PanicErr{PanicObj: i}
	}
}

func PanicToError(f func()) (err error) {
	defer func() {
		out := recover()
		if out != nil {
			err = interfaceToError(out)
		}
	}()
	f()
	return err
}
