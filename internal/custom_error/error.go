package custom_error

import (
	"fmt"
	"runtime"
)

func NewError(err error) error {
	pc, _, line, _ := runtime.Caller(1)
	fn := runtime.FuncForPC(pc).Name()
	return fmt.Errorf("%s: line %d -> %w", fn, line, err)
}
