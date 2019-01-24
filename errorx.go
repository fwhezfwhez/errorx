package errorx

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"time"
)

const (
	LcauseBy = 1 << iota
	LdateTime
	Llongfile
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
	Context     map[string]interface{}
	// upper
	ReGenerated bool
	Errors      []error
	Flag        int
}

func (e Error) String() string{
	return e.StackTraceValue()
}

func Empty() Error {
	return Error{
		E:           nil,
		StackTraces: make([]string, 0, 30),
		Context:     make(map[string]interface{}, 0),
		ReGenerated: false,
		Errors:      make([]error, 0, 30),
		Flag:        Llongfile | LcauseBy | LdateTime,
	}
}

func (e Error) Error() string {
	return e.StackTraceValue()
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
		trace := PrintStackFormat(v.Flag, file, line, v.Error())
		v.StackTraces = append(v.StackTraces, trace)
		return v
	case error:
		errorX := Empty()
		errorX.E = e
		errorX.Errors = append(errorX.Errors, e)
		_, file, line, _ := runtime.Caller(1)
		trace := PrintStackFormat(Llongfile|LcauseBy|LdateTime, file, line, v.Error())
		errorX.StackTraces = append(errorX.StackTraces, trace)
		return errorX
	}
	return New(errors.New("invalid error type,error type should be official or errorx.Error"))

}

// print error stack trace
// e.Flag only controls 'Error' type or official error which has been wrapped to 'Error'.
// e.Flag only controls stack trace inner an Error type ,rather than some stack trace which has been init by api NewFromStackTrace([]string,string)
func (e Error) PrintStackTrace() string {
	header := make([]string, 0, 10)
	if e.Flag&LdateTime > 0 {
		header = append(header, "HappenAt")
	}
	if e.Flag&Llongfile > 0 {
		header = append(header, "StackTrace")
	}
	if e.Flag&LcauseBy > 0 {
		header = append(header, "CauseBy")
	}
	fmt.Println(strings.Join(header, " | "))

	for _, v := range e.StackTraces {
		fmt.Println(v)
	}
	return strings.Join(e.StackTraces, "\n")
}

func (e Error) StackTraceValue() string {
	header := make([]string, 0, 10)
	if e.Flag&LdateTime > 0 {
		header = append(header, "HappenAt")
	}
	if e.Flag&Llongfile > 0 {
		header = append(header, "StackTrace")
	}
	if e.Flag&LcauseBy > 0 {
		header = append(header, "CauseBy")
	}
	headerStr := strings.Join(header, " | ")
	rs := make([]string, 0,len(e.StackTraces)+1)
	rs = append(rs,headerStr)
	rs = append(rs,e.StackTraces...)
	return strings.Join(rs, "\n")
}
// wrap an official error to Error type
// function do the same as New()
// New() or Wrap() depends on its semantics. mixing them is also correct.
func Wrap(e error) error {
	switch v := e.(type) {
	case Error:
		_, file, line, _ := runtime.Caller(1)
		trace := PrintStackFormat(v.Flag, file, line, v.Error())
		v.StackTraces = append(v.StackTraces, trace)
		return v
	case error:
		errorX := Empty()
		errorX.E = e
		errorX.Errors = append(errorX.Errors, e)
		_, file, line, _ := runtime.Caller(1)
		trace := PrintStackFormat(LdateTime|Llongfile|LcauseBy, file, line, v.Error())
		errorX.StackTraces = append(errorX.StackTraces, trace)
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
		trace := PrintStackFormat(v.Flag, file, line, v.Error())
		v.StackTraces = append(v.StackTraces, trace)
		return v
	case error:
		errorX := Empty()
		errorX.E = e
		errorX.Errors = append(errorX.Errors, e)
		_, file, line, _ := runtime.Caller(1)
		trace := PrintStackFormat(Llongfile|LcauseBy|LdateTime, file, line, v.Error())
		errorX.StackTraces = append(errorX.StackTraces, trace)
		return errorX
	}
	return New(errors.New("invalid error type,error type should be official or errorx.Error"))
}

// new a error from a well format string
func NewFromStringf(format string, msg ... interface{}) error {
	return NewFromString(fmt.Sprintf(format, msg...))
}

// group series of error to a single error
func GroupErrors(errors ...error) error {
	var tmp error
	var stackTrace = make([]string, 0, 10*len(errors))
	stackTrace = append(stackTrace, "######## Get a bunch of errors ########")
	defer func() {
		stackTrace = append(stackTrace, "######## /Get a bunch of errors ########")
	}()

	for i, e := range errors {
		tmp = New(e)
		stackTrace = append(stackTrace, fmt.Sprintf("\n##### error_%d #####",i))
		stackTrace = append(stackTrace, tmp.(Error).StackTraces...)
		stackTrace = append(stackTrace, fmt.Sprintf("##### /error_%d #####",i))
	}
	return NewFromStackTrace(stackTrace, "an error bunch")
}

// new an error with existed stacktrace and generate the new error with new msg
func NewFromStackTrace(stackTrace []string, msg string) error {
	e := errors.New(msg)
	v := Empty()
	v.E = e
	v.StackTraces = stackTrace
	v.Errors = append(v.Errors, e)
	_, file, line, _ := runtime.Caller(1)
	trace := PrintStackFormat(v.Flag, file, line, v.Error())
	v.StackTraces = append(v.StackTraces, trace)
	return v
}

// ReGenerate a new error and save the old
func ReGen(old error, new error) error {
	e := Empty()
	e.E = old
	e.ReGenerated = true

	switch o := old.(type) {
	case Error:
		e = o
		e.ReGenerated = true
		_, file, line, _ := runtime.Caller(1)
		trace := PrintStackFormat(e.Flag, file, line, new.Error())
		e.StackTraces = append(e.StackTraces, trace)
		e.Errors = append(e.Errors, old, new)
	case error:
		_, file, line, _ := runtime.Caller(1)
		trace := PrintStackFormat(e.Flag, file, line, new.Error())
		e.StackTraces = append(e.StackTraces, trace)
		e.Errors = append(e.Errors, old, new)
	}
	return e
}

// return an Error regardless of e's type
func MustWrap(e error) Error {
	switch v := e.(type) {
	case Error:
		return v
	case error:
		return New(v).(Error)
	}
	return Empty()
}

func PrintStackFormat(flag int, file string, line int, cause string) string {
	var formatGroup = make([]string, 0, 3)
	var formatArgs = make([]interface{}, 0, 3)
	if flag&LdateTime > 0 {
		formatGroup = append(formatGroup, "%s")
		formatArgs = append(formatArgs, time.Now().Format("2006-01-02 15:04:05"))
	}
	if flag&Llongfile > 0 {
		formatGroup = append(formatGroup, "%s")
		trace := fmt.Sprintf("%s: %d", file, line)
		formatArgs = append(formatArgs, trace)
	}
	if flag&LcauseBy > 0 {
		formatGroup = append(formatGroup, "%s")
		formatArgs = append(formatArgs, cause)
	}
	return fmt.Sprintf(strings.Join(formatGroup, " | "), formatArgs...)
}
