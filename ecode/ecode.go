package ecode

import (
	"fmt"
	"strconv"
	"sync/atomic"

	"github.com/pkg/errors"
)

var (
	_messages atomic.Value           // NOTE: stored map[string]map[int]string
	_codes    = map[int32]struct{}{} // register codes.
)

// Register register ecode message map.
func Register(cm map[int32]string) {

	for k, v := range defaultCodeMap {
		if _, ok := cm[k]; !ok {
			cm[k] = v
		}
	}

	_messages.Store(cm)
}

// New new a ecode.Codes by int value.
// NOTE: ecode must unique in global, the New will check repeat and then panic.
func New(e int32) Code {
	if e <= 0 {
		panic("business ecode must greater than zero")
	}
	return add(e)
}

func add(e int32) Code {
	if _, ok := _codes[e]; ok {
		panic(fmt.Sprintf("ecode: %d already exist", e))
	}
	_codes[e] = struct{}{}
	return Int32(e)
}

// Codes ecode error interface which has a code & message.
type Codes interface {
	// sometimes Error return Code in string form
	// NOTE: don't use Error in monitor report even it also work for now
	Error() string
	// Code get error code.
	Code() int32
	// Message get code message.
	Message() string

	CodeAndMessage() (int32, string)
	//Detail get error detail,it may be nil.
	Details() []interface{}
	// Equal for compatible.
	// Deprecated: please use ecode.EqualError.
	Equal(error) bool
}

// A Code is an int32 error code spec.
type Code int32

func (e Code) Error() string {
	return strconv.FormatInt(int64(e), 10)
}

// Code return error code
func (e Code) Code() int32 { return int32(e) }

// Message return error message
func (e Code) Message() string {
	if cm, ok := _messages.Load().(map[int32]string); ok {
		if msg, ok := cm[e.Code()]; ok {
			return msg
		}
	}
	return e.Error()
}

// CodeAndMessage .
func (e Code) CodeAndMessage() (int32, string) {
	return e.Code(), e.Message()
}

// Details return details.
func (e Code) Details() []interface{} { return nil }

// Equal for compatible.
// Deprecated: please use ecode.EqualError.
func (e Code) Equal(err error) bool { return EqualError(e, err) }

// Int32 parse code int to error.
func Int32(i int32) Code { return Code(i) }

// String parse code string to error.
func String(e string) Code {
	if e == "" {
		return OK
	}
	// try error string
	i, err := strconv.Atoi(e)
	if err != nil {
		return ErrServer
	}
	return Code(i)
}

// Cause cause from error to ecode.
func Cause(e error) Codes {
	if e == nil {
		return OK
	}
	ec, ok := errors.Cause(e).(Codes)
	if ok {
		return ec
	}
	return String(e.Error())
}

// Equal equal a and b by code int.
func Equal(a, b Codes) bool {
	if a == nil {
		a = OK
	}
	if b == nil {
		b = OK
	}
	return a.Code() == b.Code()
}

// EqualError equal error
func EqualError(code Codes, err error) bool {
	return Cause(err).Code() == code.Code()
}
