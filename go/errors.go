package sdk

import "fmt"

// Error represents an API error response.
type Error struct {
	_         struct{} // callers must name fields
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
