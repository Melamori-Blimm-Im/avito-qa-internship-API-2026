package utils

import (
	"context"
	"errors"
)

// IsClientTimeout возвращает true, если err — таймаут http.Client, context.DeadlineExceeded или иной сетевой таймаут.
func IsClientTimeout(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}
	type timeouter interface{ Timeout() bool }
	var te timeouter
	return errors.As(err, &te) && te.Timeout()
}
