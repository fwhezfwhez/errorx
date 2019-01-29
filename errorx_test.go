package errorx

import (
	"errors"
	"fmt"
	"testing"
)

type Error2 struct {
	// basic
	E           error
	StackTraces []string

	// upper
	ReGenerated bool
	Errors      []error
}

// test the stacktrace of an error
func TestNew(t *testing.T) {
	// starts from a string msg
	fun5 := func() error {
		return NewFromString("made an error")
	}

	fun3 := func() error {
		e := fun5()
		if e != nil {
			return Wrap(e)
		}
		return nil
	}
	fun2 := func() error {
		e := fun3()
		if e != nil {
			return Wrap(e)
		}
		return nil
	}

	fun1 := func() error {
		e := fun2()
		if e != nil {
			return New(e)
		}
		return nil
	}

	e := fun1()

	if e != nil {
		t.Log(e.Error())
		//fmt.Println(e.(Error).PrintStackTrace())
		e.(Error).PrintStackTrace()
	}
}

//test new an error from a string msg
func TestNewFromString(t *testing.T) {
	fun2 := func() error {
		return NewFromString("an error happen")
	}

	fun3 := func() error {
		if e := fun2(); e != nil {
			return New(e)
		}
		return nil
	}

	er := fun3()
	if er != nil {
		fmt.Println("origin error:", er.Error())
		er.(Error).PrintStackTrace()
	}
}

//test new an error with existed stacktrace and generate the new error with new msg
func TestNewFromStackTrace(t *testing.T) {
	fun2 := func() error {
		return NewFromStackTrace([]string{
			"G:/go_workspace/GOPATH/src/errorX/errorx_test.go: 49 | an error happen",
			"G:/go_workspace/GOPATH/src/errorX/errorx_test.go: 43 | an error happen",
		}, "inner service error")
	}

	fun3 := func() error {
		if e := fun2(); e != nil {
			return New(e)
		}
		return nil
	}

	er := fun3()
	if er != nil {
		fmt.Println("origin error:", er.Error())
		er.(Error).PrintStackTrace()
	}
}

// test regenerate a new error
func TestReGen(t *testing.T) {
	// start from an official error
	fun1 := func() error {
		return errors.New("old official error")
	}
	fun2 := func() error {
		if e := fun1(); e != nil {
			return New(e)
		}
		return nil
	}

	fun3 := func() error {
		if e := fun2(); e != nil {
			return New(e)
		}
		return nil
	}
	fun4 := func() error {
		if e := fun3(); e != nil {
			return ReGen(e, errors.New("inner service error"))
		}
		return nil
	}

	er := fun4()
	if er != nil {
		fmt.Println("origin error:", er.Error())
		er.(Error).PrintStackTrace()
	}
}

func TestMustWrap(t *testing.T) {
	//fmt.Println(MustWrap(errors.New("offial error")))
	//fmt.Println(MustWrap(Empty()))
	//fmt.Println(Empty())
	//e := Error{
	//	E:           nil,
	//	StackTraces: make([]string, 0, 30),
	//	ReGenerated: false,
	//	Errors:      make([]error, 0, 30),
	//}
	//fmt.Println(e)
}

//test flagFormat
func TestPrintStackFormat(t *testing.T) {
	rs := PrintStackFormat(LdateTime|LcauseBy|Llongfile, "main.go", 12, "an error happen")
	fmt.Println(rs)
}

func TestErrorGroup(t *testing.T) {
	t.Fail()
	return
	var length = 10
	var errors = make([]error, 0, length)
	for i := 0; i < length; i++ {
		errors = append(errors, NewFromStringf("error_%d", i+1))
	}
	e := GroupErrors(errors...)
	fmt.Println(e)
}
