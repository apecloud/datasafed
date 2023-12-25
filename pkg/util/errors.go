package util

import (
	"fmt"
	"strings"
)

func WrappedErrOrNil(err error, format string, args ...interface{}) error {
	if err == nil {
		return nil
	}
	msg := fmt.Sprintf(format, args...)
	if !strings.Contains(msg, "%w") {
		msg = msg + ", error: %w"
	}
	return fmt.Errorf(msg, err)
}
