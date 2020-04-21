package errorx

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"runtime"
	"strconv"
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
	// Each wrap will add index 1
	index int

	// basic
	E           error
	StackTraces []string

	// do not use context, it's tracking bug on development
	Context map[string]interface{}
	Header  map[string][]string

	// upper
	ReGenerated bool
	Errors      []error
	Flag        int
	Keyword     string
}

func (e Error) String() string {
	return e.Error()
}

// set header for an error obj
// SetHeader is safe whether e.Header is nil or not
func (e *Error) SetHeader(key string, value string) {
	if e.Header == nil {
		e.Header = make(map[string][]string, 0)
	}
	if e.Header[key] == nil {
		e.Header[key] = make([]string, 0, 10)
	}
	if len(e.Header[key]) == 0 {
		e.Header[key] = append(e.Header[key], value)
	} else {
		e.Header[key][0] = value
	}
}

// GetHeader is safe wheter header is nil
func (e Error) GetHeader(key string) string {
	if e.Header == nil {
		return ""
	}
	if e.Header[key] == nil || len(e.Header[key]) == 0 {
		return ""
	}
	return e.Header[key][0]
}
func Empty() Error {
	return Error{
		index:       0,
		E:           nil,
		StackTraces: make([]string, 0, 30),
		Context:     make(map[string]interface{}, 0),
		ReGenerated: false,
		Errors:      make([]error, 0, 30),
		Flag:        Llongfile | LcauseBy | LdateTime,
	}
}

func (e Error) Error() string {
	var header = ""
	if len(e.Header) != 0 {
		buf, _ := json.MarshalIndent(e.Header, "  ", "  ")
		header += fmt.Sprintf("header:\n%s\n", buf)
	}
	rs := fmt.Sprintf("%s%s\n", header, e.StackTraceValue())

	if len(e.Context) != 0 {
		buf, _ := json.MarshalIndent(e.Context, "  ", "  ")
		rs += fmt.Sprintf("context:\n%s\n", buf)
	}
	return rs
}
func (e Error) BasicError() string {
	if e.E != nil {
		return e.E.Error()
	}
	return ""
}

// return its stacktrace
func (e Error) StackTrace() string {
	rs := "\n"
	for _, v := range e.StackTraces {
		rs = rs + v + "\n"
	}
	return rs
}

// New an error with header info
func NewWithHeader(e error, header map[string]interface{}) error {
	if e == nil {
		return nil
	}

	er := New(e).(Error)
	for k, v := range header {
		er.SetHeader(k, ToString(v))
	}
	return er
}

func NewWithAttach(e error, msg interface{}) error {
	if e == nil {
		return nil
	}

	msg = ToString(msg)
	er := New(e).(Error)
	er.SetHeader("attach", msg.(string))
	return er
}

// New a error
func New(e error) error {
	if e == nil {
		return nil
	}
	switch v := e.(type) {
	case Error:
		_, file, line, _ := runtime.Caller(1)
		trace := PrintStackFormat(v.Flag, file, line, v.BasicError())
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
	//header := make([]string, 0, 10)
	//if e.Flag&LdateTime > 0 {
	//	header = append(header, "HappenAt")
	//}
	//if e.Flag&Llongfile > 0 {
	//	header = append(header, "StackTrace")
	//}
	//if e.Flag&LcauseBy > 0 {
	//	header = append(header, "CauseBy")
	//}
	//fmt.Println(strings.Join(header, " | "))

	for _, v := range e.StackTraces {
		fmt.Println(v)
	}
	return strings.Join(e.StackTraces, "\n")
}

func (e Error) StackTraceValue() string {
	//header := make([]string, 0, 10)
	//if e.Flag&LdateTime > 0 {
	//	header = append(header, "HappenAt")
	//}
	//if e.Flag&Llongfile > 0 {
	//	header = append(header, "StackTrace")
	//}
	//if e.Flag&LcauseBy > 0 {
	//	header = append(header, "CauseBy")
	//}
	//headerStr := strings.Join(header, " | ")
	rs := make([]string, 0, len(e.StackTraces)+1)
	// rs = append(rs, headerStr)
	rs = append(rs, e.StackTraces...)
	return strings.Join(rs, "\n")
}

// wrap an official error to Error type
// function do the same as New()
// New() or Wrap() depends on its semantics. mixing them is also correct.
func Wrap(e error) error {
	if e == nil {
		return nil
	}
	switch v := e.(type) {
	case Error:
		_, file, line, _ := runtime.Caller(1)
		trace := PrintStackFormat(v.Flag, file, line, v.BasicError())
		v.StackTraces = append(v.StackTraces, trace)
		v.index++
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

// new an error from string
func NewFromString(msg string) error {
	e := errors.New(msg)
	switch v := e.(type) {
	case Error:
		_, file, line, _ := runtime.Caller(1)
		trace := PrintStackFormat(v.Flag, file, line, v.BasicError())
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

func NewFromStringWithDepth(msg string, depth int) error {
	e := errors.New(msg)
	switch v := e.(type) {
	case Error:
		_, file, line, _ := runtime.Caller(depth)
		trace := PrintStackFormat(v.Flag, file, line, v.BasicError())
		v.StackTraces = append(v.StackTraces, trace)
		return v
	case error:
		errorX := Empty()
		errorX.E = e
		errorX.Errors = append(errorX.Errors, e)
		_, file, line, _ := runtime.Caller(depth)
		trace := PrintStackFormat(Llongfile|LcauseBy|LdateTime, file, line, v.Error())
		errorX.StackTraces = append(errorX.StackTraces, trace)
		return errorX
	}
	return New(errors.New("invalid error type,error type should be official or errorx.Error"))
}

func newFromStringWithDepth(msg string, depth int) error {
	e := errors.New(msg)
	switch v := e.(type) {
	case Error:
		_, file, line, _ := runtime.Caller(depth)
		trace := PrintStackFormat(v.Flag, file, line, v.BasicError())
		v.StackTraces = append(v.StackTraces, trace)
		return v
	case error:
		errorX := Empty()
		errorX.E = e
		errorX.Errors = append(errorX.Errors, e)
		_, file, line, _ := runtime.Caller(depth)
		trace := PrintStackFormat(Llongfile|LcauseBy|LdateTime, file, line, v.Error())
		errorX.StackTraces = append(errorX.StackTraces, trace)
		return errorX
	}
	return New(errors.New("invalid error type,error type should be official or errorx.Error"))
}

// new an error from string with header
func NewFromStringWithHeader(msg string, header map[string]interface{}) error {
	er := NewFromString(msg).(Error)
	for k, v := range header {
		er.SetHeader(k, ToString(v))
	}
	return er
}

// new a error from a well format string with header
func NewFromStringWithHeaderf(format string, msg string, header map[string]interface{}) error {
	er := NewFromStringf(format, msg).(Error)
	for k, v := range header {
		er.SetHeader(k, ToString(v))
	}
	return er
}

// new an error from string with header
func NewFromStringWithAttach(msg string, attach interface{}) error {
	er := NewFromString(msg).(Error)
	er.SetHeader("attach", ToString(attach))
	return er
}

// new an error from  well format string with header
func NewFromStringWithAttachf(format string, msg string, attach interface{}) error {
	er := NewFromStringf(format, msg).(Error)
	er.SetHeader("attach", ToString(attach))
	return er
}

// new a error from a well format string
func NewFromStringf(format string, msg ...interface{}) error {
	return newFromStringWithDepth(fmt.Sprintf(format, msg...), 2)
}

// new a error from a error  with numeric params
func NewWithParam(e error, params ...interface{}) error {
	if e == nil {
		return nil
	}

	if len(params) == 0 {
		return Wrap(e)
	}
	type ErrWithParam struct {
		ErrMsg string      `json:"error"`
		Params interface{} `json:"params"`
	}
	var errorMsg string
	switch v := e.(type) {
	case Error:
		errorMsg = fmt.Sprintf("\n" + strings.Join(v.StackTraces, "\n"))
	case error:
		errorMsg = fmt.Sprintf(v.Error())
	}

	// record param
	ep := ErrWithParam{}
	ep.ErrMsg = errorMsg
	ep.Params = params
	buf, _ := json.MarshalIndent(ep, "", "  ")
	return NewFromString(fmt.Sprintf("%s", buf))
}
func NewFromStringWithParam(msg string, params ...interface{}) error {
	if len(params) == 0 {
		return NewFromString(msg)
	}
	type ErrWithParam struct {
		ErrMsg string      `json:"error"`
		Params interface{} `json:"params"`
	}
	// record param
	ep := ErrWithParam{}
	ep.ErrMsg = msg
	ep.Params = params
	buf, _ := json.MarshalIndent(ep, "", "  ")
	return NewFromString(fmt.Sprintf("%s", buf))
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
		if e == nil {
			continue
		}
		tmp = New(e)
		stackTrace = append(stackTrace, fmt.Sprintf("\n##### error_%d #####", i))
		stackTrace = append(stackTrace, tmp.(Error).StackTraces...)
		stackTrace = append(stackTrace, fmt.Sprintf("##### /error_%d #####", i))
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

func WrapContext(e error, ctx map[string]interface{}) error {
	if e == nil {
		return nil
	}
	switch v := e.(type) {
	case Error:
		_, file, line, _ := runtime.Caller(1)
		trace := PrintStackFormat(v.Flag, file, line, v.BasicError())
		v.StackTraces = append(v.StackTraces, trace)
		if v.Context == nil || len(v.Context) == 0 {
			v.Context = ctx
		} else {
			for i, va := range ctx {
				_, ok := v.Context[i]
				if !ok {
					v.Context[i] = va
				} else {
					v.Context[fmt.Sprintf("%s_%d", i, v.index)] = va
				}
			}
		}
		return v
	case error:
		errorX := Empty()
		errorX.E = e
		errorX.Errors = append(errorX.Errors, e)
		_, file, line, _ := runtime.Caller(1)
		trace := PrintStackFormat(LdateTime|Llongfile|LcauseBy, file, line, v.Error())
		errorX.StackTraces = append(errorX.StackTraces, trace)
		if len(ctx) != 0 {
			errorX.Context = ctx
		}
		return errorX
	}
	return Wrap(errors.New("invalid error type,error type should be official or errorx.Error"))
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
		trace := fmt.Sprintf("%s:%d", file, line)
		formatArgs = append(formatArgs, trace)
	}
	if flag&LcauseBy > 0 {
		formatGroup = append(formatGroup, "%s")
		formatArgs = append(formatArgs, cause)
	}
	return fmt.Sprintf(strings.Join(formatGroup, " | "), formatArgs...)
}

// generate an error key word.
// key word is used to help save errors in database
// It's suggested to set unique(date, keyword), when error with same keyword in a day,database only saves field 'times'
// rather than another error record
func (e Error) GenerateKeyword() string {
	arr := strings.Split((e).StackTraces[len(e.StackTraces)-1], "|")
	core := strings.TrimSpace(arr[len(arr)-1])
	return generateKeyWord(core)
}

// generate key word to an error type
func GenerateKeyword(e error) string {
	switch v := e.(type) {
	case Error:
		return v.GenerateKeyword()
	case error:
		return generateKeyWord(e.Error())
	}
	return generateKeyWord(e.Error())
}

// generate key word ruled:
// time out from mysql database     -> tofmd
// connection panic from exception  -> cpfe
func generateKeyWord(in string) string {
	arr := Split(in, " ")
	var result string
	for _, v := range arr {
		result += string(v[0])
	}
	return result
}

// Split better strings.Split，对  a,,,,,,,b,,c     以","进行切割成[a,b,c]
func Split(s string, sub string) []string {
	var rs = make([]string, 0, 20)
	tmp := ""
	Split2(s, sub, &tmp, &rs)
	return rs
}

// Split2 附属于Split，可独立使用
func Split2(s string, sub string, tmp *string, rs *[]string) {
	s = strings.Trim(s, sub)
	if !strings.Contains(s, sub) {
		*tmp = s
		*rs = append(*rs, *tmp)
		return
	}
	for i := range s {
		if string(s[i]) == sub {
			*tmp = s[:i]
			*rs = append(*rs, *tmp)
			s = s[i+1:]
			Split2(s, sub, tmp, rs)
			return
		}
	}
}

func ToString(arg interface{}) string {
	tmp := reflect.Indirect(reflect.ValueOf(arg)).Interface()
	switch v := tmp.(type) {
	case int:
		return strconv.Itoa(v)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case string:
		return v
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case time.Time:
		return v.Format("2006-01-02 15:04:05")
	case fmt.Stringer:
		return v.String()
	default:
		return ""
	}
}
