package errorx

import (
	"errors"
	"fmt"
	"runtime"
)

// an Error instance wraps an official error and storage its stackTrace.
// E and StackTraces work for the basic server
// When tries to wrap the origin error to another, logic does like:
//			if er := f(); er!=nil{
//			    log.Println(er.Error())
//				return errors.New("inner service error")
//			}
//
// 'ReGenerated' should be set true and different errors should be saved in 'Errors',and the 'stackTrace' should include
// all stackTraces above.But in this case,'Error.E' only point to the origin one
// It's ok to storage both official error and 'Error'
type Error struct {
	// basic
	E           error
	StackTraces []string

	// upper
	ReGenerated bool
	Errors      []error
}

func Empty() Error {
	return Error{
		E:           nil,
		StackTraces: make([]string, 0, 30),
		ReGenerated: false,
		Errors:      make([]error, 0, 30),
	}
}

func (e Error) Error() string {
	return e.E.Error()
}

// return its stacktrace
func (e Error) StackTrace() string {
	rs := "\n"
	for _, v := range e.StackTraces {
		rs = rs + v + "\n"
	}
	return rs
}

// New a error
func New(e error) error {
	switch v := e.(type) {
	case Error:
		_, file, line, _ := runtime.Caller(1)
		v.StackTraces = append(v.StackTraces, fmt.Sprintf("%s: %d | %s", file, line, e.Error()))
		return v
	case error:
		errorX := Error{
			E:           e,
			StackTraces: make([]string, 0, 30),
			Errors:      append(make([]error, 0, 30), e),
			ReGenerated: false,
		}
		_, file, line, _ := runtime.Caller(1)
		errorX.StackTraces = append(errorX.StackTraces, fmt.Sprintf("%s: %d | %s", file, line, e.Error()))
		return errorX
	}
	return New(errors.New("invalid error type,error type should be official or errorx.Error"))

}

// print error stack trace
func (e Error) PrintStackTrace() {
	fmt.Println(fmt.Sprintf("%s | %s", "StackTrace", "CausedBy"))
	for _, v := range e.StackTraces {
		fmt.Println(v)
	}
}

// wrap an official error to Error type
// function do the same as New()
// New() or Wrap() depends on its semantics. mixing them is also correct.
func Wrap(e error) error {
	switch v := e.(type) {
	case Error:
		_, file, line, _ := runtime.Caller(1)
		v.StackTraces = append(v.StackTraces, fmt.Sprintf("%s: %d | %s", file, line, e.Error()))
		return v
	case error:
		errorX := Error{
			E:           e,
			StackTraces: make([]string, 0, 30),
			Errors:      append(make([]error, 0, 30), e),
			ReGenerated: false,
		}
		_, file, line, _ := runtime.Caller(1)
		errorX.StackTraces = append(errorX.StackTraces, fmt.Sprintf("%s: %d | %s", file, line, e.Error()))
		return errorX
	}
	return Wrap(errors.New("invalid error type,error type should be official or errorx.Error"))
}

// new a error from string
func NewFromString(msg string) error {
	e := errors.New(msg)
	switch v := e.(type) {
	case Error:
		_, file, line, _ := runtime.Caller(1)
		v.StackTraces = append(v.StackTraces, fmt.Sprintf("%s: %d | %s", file, line, e.Error()))
		return v
	case error:
		errorX := Error{
			E:           e,
			StackTraces: make([]string, 0, 30),
			Errors:      append(make([]error, 0, 30), e),
			ReGenerated: false,
		}
		_, file, line, _ := runtime.Caller(1)
		errorX.StackTraces = append(errorX.StackTraces, fmt.Sprintf("%s: %d | %s", file, line, e.Error()))
		return errorX
	}
	return New(errors.New("invalid error type,error type should be official or errorx.Error"))
}

// new an error with existed stacktrace and generate the new error with new msg
func NewFromStackTrace(stackTrace []string, msg string) error {
	e := errors.New(msg)
	v := Error{
		E:           e,
		StackTraces: stackTrace,
		Errors:      append(make([]error, 0, 30), e),
		ReGenerated: false,
	}
	_, file, line, _ := runtime.Caller(1)
	trace := fmt.Sprintf("%s: %d | %s", file, line, e.Error())
	v.StackTraces = append(v.StackTraces, trace)
	return v
}

// ReGenerate a new error and save the old
func ReGen(old error, new error) error {
	e := Error{
		E:           old,
		StackTraces: make([]string, 0, 30),
		Errors:      make([]error, 0, 30),
		ReGenerated: true,
	}

	switch o := old.(type) {
	case Error:
		e = o
		e.ReGenerated = true
		_, file, line, _ := runtime.Caller(1)
		trace := fmt.Sprintf("%s: %d | %s", file, line, new.Error())
		e.StackTraces = append(e.StackTraces, trace)
		e.Errors = append(e.Errors, old, new)
	case error:
		_, file, line, _ := runtime.Caller(1)
		trace := fmt.Sprintf("%s: %d | %s", file, line, new.Error())
		e.StackTraces = append(e.StackTraces, trace)
		e.Errors = append(e.Errors, old, new)
	}
	return e
}

// return an Error regardless of e's type
func MustWrap(e error) Error {
	switch v:=e.(type){
	case Error:
		return v
	case error:
		return New(v).(Error)
	}
	return Empty()
}
