package domain

import "fmt"

type errcode int64

const (
	Undefined errcode = iota
	Internal
	Security
)

func (s errcode) String() string {
	switch s {
	case Undefined:
		return "undefined"
	case Internal:
		return "internal"
	case Security:
		return "security"
	}
	return "unknown"
}

type Error struct {
	Code    errcode
	Message string // Human-readable error message.
	Source  error  // Machine-readable error message.
}

func (e *Error) Error() string {
	return fmt.Sprintf("application error: code=%s message=%s", e.Code, e.Message)
}

func Errorf(source error, code errcode, format string, args ...interface{}) *Error {
	return &Error{
		Code:    code,
		Message: fmt.Sprintf(format, args...),
		Source:  source,
	}
}
