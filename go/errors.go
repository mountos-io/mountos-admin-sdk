package sdk

import "fmt"

// Error represents an API error response.
type Error struct {
	Message   string
	Status    int
	ErrorCode int
}

func (e *Error) Error() string {
	if e.ErrorCode != 0 {
		return fmt.Sprintf("mountos: %s (status=%d, code=%d)", e.Message, e.Status, e.ErrorCode)
	}
	return fmt.Sprintf("mountos: %s (status=%d)", e.Message, e.Status)
}
