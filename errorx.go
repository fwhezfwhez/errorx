package errorx

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
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
// When tries to wrap the origin error to another, logic does like:
//
//	if er := f(); er!=nil{
//	    log.Println(er.Error())
//		return errors.New("inner service error")
//	}
//
// 'ReGenerated' should be set true and different errors should be saved in 'Errors',and the 'stackTrace' should include
// all stackTraces above.But in this case,'Error.E' only point to the origin one
// It's ok to storage both official error and 'Error'
type Error struct {
	// Each wrap will add index 1
	index int

	// core error // saves origin error

	// basic
	E             error
	StackTraces   []string    // Deprecated,保留了字段防止外部应用直接使用引起无法编译问题
	stackTracesV2 []stackline // 新版堆栈路径

	isServiceErr   bool
	serviceErrcode int
	serviceErrmsg  string

	// do not use context, it's tracking bug on development
	Context map[string]interface{}
	Header  map[string][]string

	// upper
	ReGenerated bool
	Errors      []error
	Flag        int
	Keyword     string
}

func (e *Error) wrapStackLine(desc string) {

	trace := stackDepth3()
	if len(e.stackTracesV2) == 0 {
		e.stackTracesV2 = make([]stackline, 0, 10)
	}

	e.stackTracesV2 = append([]stackline{{
		trace: trace,
		desc:  desc,
	}}, e.stackTracesV2...)
}

func (e *Error) wrapStackLineWithTrace(trace string, desc string) {
	if len(e.stackTracesV2) == 0 {
		e.stackTracesV2 = make([]stackline, 0, 10)
	}

	e.stackTracesV2 = append([]stackline{{
		trace: trace,
		desc:  desc,
	}}, e.stackTracesV2...)
}

func (e Error) Stack() []string {
	var rs = make([]string, 0, 10)

	for _, v := range e.stackTracesV2 {
		tmp := fmt.Sprintf("%s %s", v.trace, v.desc)
		rs = append(rs, tmp)
	}

	return rs
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
	return empty()
}

func empty() Error {
	return Error{
		index:         0,
		E:             nil,
		StackTraces:   make([]string, 0, 30),
		stackTracesV2: make([]stackline, 0, 10),
		Context:       make(map[string]interface{}, 0),
		ReGenerated:   false,
		Errors:        make([]error, 0, 30),
		Flag:          Llongfile | LcauseBy | LdateTime,
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
	return e.StackTraceValue()
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

	tmp := empty()
	tmp.E = e
	tmp.wrapStackLine(e.Error())
	return tmp
}

// print error stack trace
// e.Flag only controls 'Error' type or official error which has been wrapped to 'Error'.
// e.Flag only controls stack trace inner an Error type ,rather than some stack trace which has been init by api NewFromStackTrace([]string,string)
func (e Error) PrintStackTrace() string {
	return e.StackTraceValue()
}

func (e Error) StackTraceValue() string {

	rs := make([]string, 0, 10)

	for _, v := range e.stackTracesV2 {
		tmp := fmt.Sprintf("%s %s", v.trace, v.desc)
		rs = append(rs, tmp)
	}

	return strings.Join(rs, " | ")
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
		v.wrapStackLine("")
		v.index++
		return v
	case ServiceError:
		errorX := Empty()
		errorX.E = e
		errorX.isServiceErr = true
		errorX.serviceErrcode = v.Errcode
		errorX.serviceErrmsg = v.Errmsg

		errorX.wrapStackLine(v.Errmsg)
		return errorX
	case error:
		errorX := empty()
		errorX.E = e
		errorX.Errors = append(errorX.Errors, e)

		errorX.wrapStackLine(e.Error())
		return errorX
	}
	return Wrap(errors.New("invalid error type,error type should be official or errorx.Error"))
}

func NewFromStringWithDepth(msg string, depth int) error {
	e := empty()
	e.E = fmt.Errorf("%s", msg)
	e.wrapStackLine(msg)

	return e
}

// new an error from string with header
func NewFromStringWithHeader(msg string, header map[string]interface{}) error {
	var tmp = make([]string, 0, 10)

	for k, v := range header {
		pair := fmt.Sprintf("%s=%v", k, v)
		tmp = append(tmp, pair)
	}

	headerInfo := strings.Join(tmp, " ")

	info := fmt.Sprintf("%s err=%s", headerInfo, msg)

	e := empty()
	e.E = fmt.Errorf("%s", msg)
	e.wrapStackLine(info)

	return e
}

// new a error from a well format string with header
func NewFromStringWithHeaderf(format string, msg string, header map[string]interface{}) error {

	var tmp = make([]string, 0, 10)

	for k, v := range header {
		pair := fmt.Sprintf("%s=%v", k, v)
		tmp = append(tmp, pair)
	}

	headerInfo := strings.Join(tmp, " ")

	info := fmt.Sprintf("%s err=%s", headerInfo, msg)

	e := empty()
	e.E = fmt.Errorf("%s", msg)
	e.wrapStackLine(info)

	return e
}

// Deprecated: using NewFromStringf
func NewFromStringWithAttach(msg string, attach interface{}) error {
	info := fmt.Sprintf("err=%s attach=%v", msg, attach)

	e := empty()
	e.E = fmt.Errorf("%s", msg)
	e.wrapStackLine(info)
	return e
}

// Deprecated: using NewFromStringf
func NewFromStringWithAttachf(format string, msg string, attach interface{}) error {
	info := fmt.Sprintf("err=%s attach=%v", msg, attach)

	e := empty()
	e.E = fmt.Errorf("%s", msg)
	e.wrapStackLine(info)
	return e
}

// new an error from string
func NewFromString(msg string) error {
	v := empty()
	v.E = fmt.Errorf("%s", msg)
	v.wrapStackLine(msg)
	return v
}

// new a error from a well format string
func NewFromStringf(format string, args ...interface{}) error {
	v := findErr(args...)
	v.wrapStackLine(fmt.Sprintf(format, args...))
	return v
}

func findErr(args ...interface{}) Error {
	var e Error
	if len(args) > 0 {
		for _, v := range args {
			if tmp, ok := v.(Error); ok {
				e = tmp
				// 找到了底层错误，则返回该错误
				return e
			}
		}

		return empty()

	}
	return empty()

}

// Deprecated: using WrapContext
func NewWithParam(e error, params ...interface{}) error {
	if e == nil {
		return nil
	}
	if len(params) == 0 {
		tmp := empty()
		tmp.E = e
		tmp.wrapStackLine(e.Error())

		return tmp
	}

	var infoarr = make([]string, 0, 10)
	for i, v := range params {
		pair := fmt.Sprintf("param%d=%v", i, v)
		infoarr = append(infoarr, pair)
	}

	paraminfo := strings.Join(infoarr, " ")

	tmp := empty()
	tmp.E = e
	tmp.wrapStackLine(fmt.Sprintf("%s %s", e.Error(), paraminfo))

	return tmp

}
func NewFromStringWithParam(msg string, params ...interface{}) error {
	if len(params) == 0 {
		tmp := empty()
		tmp.E = fmt.Errorf("%s", msg)
		tmp.wrapStackLine(msg)

		return tmp
	}

	var infoarr = make([]string, 0, 10)
	for i, v := range params {
		pair := fmt.Sprintf("param%d=%v", i, v)
		infoarr = append(infoarr, pair)
	}

	paraminfo := strings.Join(infoarr, " ")

	tmp := empty()
	tmp.E = fmt.Errorf("%s", msg)
	tmp.wrapStackLine(fmt.Sprintf("%s %s", msg, paraminfo))

	return tmp
}

// group series of error to a single error
func GroupErrors(es ...error) error {

	tmp := empty()

	infoarr := make([]string, 0, 10)
	for i, v := range es {
		if v == nil {
			continue
		}

		pair := fmt.Sprintf("error_%d=%s", i, v.Error())

		infoarr = append(infoarr, pair)
	}

	tmp.wrapStackLine(fmt.Sprintf("Group Errors %s", strings.Join(infoarr, " ")))

	return tmp
}

// new an error with existed stacktrace and generate the new error with new msg
func NewFromStackTrace(stackTrace []string, msg string) error {
	e := empty()

	for _, v := range stackTrace {
		e.wrapStackLineWithTrace(v, msg)
	}

	return e
}

func WrapContext(e error, ctx map[string]interface{}) error {
	if e == nil {
		return nil
	}

	var infoparams = make([]string, 0, 10)
	for k, v := range ctx {
		argone := fmt.Sprintf("%s=%v", k, v)
		infoparams = append(infoparams, argone)
	}
	ctxinfo := strings.Join(infoparams, " ")

	switch v := e.(type) {
	case Error:
		v.wrapStackLine(ctxinfo)
		v.index++
		return v
	case ServiceError:
		errorX := Empty()
		errorX.E = e
		errorX.isServiceErr = true
		errorX.serviceErrcode = v.Errcode
		errorX.serviceErrmsg = v.Errmsg

		errorX.wrapStackLine(fmt.Sprintf("errno=%d errmsg=%s %s", v.Errcode, v.Errmsg, ctxinfo))
		return errorX
	case error:
		errorX := empty()
		errorX.E = e
		errorX.Errors = append(errorX.Errors, e)
		errorX.wrapStackLine(fmt.Sprintf("%s %s", e.Error(), ctxinfo))
		return errorX
	}
	return Wrap(errors.New("invalid error type,error type should be official or errorx.Error"))
}

// Decrecated
func ReGen(old error, new error) error {
	return new
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

// generate an error key word.
// key word is used to help save errors in database
// It's suggested to set unique(date, keyword), when error with same keyword in a day,database only saves field 'times'
// rather than another error record
func (e Error) GenerateKeyword() string {
	arr := e.stackTracesV2
	if len(arr) == 0 {
		return "no stack"
	}
	core := strings.TrimSpace(arr[0].trace)
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

func IsError(src error, dest error) (string, bool) {
	if src == nil {
		return "", false
	}
	if dest == nil {
		return "", false
	}

	switch v := src.(type) {
	case Error:
		if v.E == nil {
			return "", false
		}
		if v.E.Error() == dest.Error() {
			return dest.Error(), true
		}
	case error:
		if v.Error() == dest.Error() {
			return dest.Error(), true
		}
	}
	return "", false
}
