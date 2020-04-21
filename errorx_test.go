package errorx

import (
	"errors"
	"fmt"
	"runtime/debug"
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
	var length = 10
	var errors = make([]error, 0, length)
	for i := 0; i < length; i++ {
		errors = append(errors, NewFromStringf("error_%d", i+1))
	}
	e := GroupErrors(errors...)
	fmt.Println(e)
}

func TestNewFromParam(t *testing.T) {
	fmt.Println(NewFromStringWithParam("no record found", struct{ Name string }{"ft"}, struct{ Age int }{9}).Error())
}

func TestError_GenerateKeyword(t *testing.T) {
	result := generateKeyWord("hello kitty beautiful")
	fmt.Println(result)
	if result != "hkb" {
		t.Fail()
		return
	}
	result = generateKeyWord("  hello    kitty   beautiful   ")
	fmt.Println(result)
	if result != "hkb" {
		t.Fail()
		return
	}

	e := NewFromString("time out from mysql")
	// result = e.(Error).GenerateKeyword()
	result = GenerateKeyword(e)

	fmt.Println(result)
	if result != "tofm" {
		t.Fail()
		return
	}

}

func TestHeader(t *testing.T) {
	fmt.Println(Wrap(nil))
	// er:= NewWithHeader(NewFromString("mysql time out"), map[string]interface{}{
	// 	"api": "/game/pay/order",
	// 	"user_id": 330392,
	// })
	// fmt.Println(er.Error())
	//
	er := NewWithAttach(NewFromString("mysql time out"), string(debug.Stack()))
	fmt.Println(er.Error())

	//er = NewFromStringWithHeaderf("user_id '%s' server error", "30939", map[string]interface{}{
	//	"api":"/game/pay/order/",
	//})
	//fmt.Println(er.Error())
	//
	//er = NewFromStringWithHeader("user_id '30939' server error", map[string]interface{}{
	//	"api":"/game/pay/order/",
	//})
	////fmt.Println(er.Error())
	//
	////er = NewFromStringWithAttach("user_id '30939' server error", "test")
	////fmt.Println(er.Error())
	//
	//er = NewWithHeader(er, map[string]interface{}{
	//	"req": "111",
	//})
	//fmt.Println(er.Error())
}

func TestMMM(t *testing.T) {
	// Wrap(Wrap(errors.New("hello"))).(Error).PrintStackTrace()
	fmt.Println(NewWithAttach(NewFromString("hello"), "attach"))
}

func TestWrapContext(t *testing.T) {
	fmt.Println(tmp().Error())
}

func TestWrap(t *testing.T) {
	fmt.Println(tmpContext().Error())
	fmt.Println(tmpWrap().Error())
}

func tmp() error {
	// return NewFromStringf("nil return")
	return Wrap(tmpContext())
}

func tmpContext() error {
	return WrapContext(fmt.Errorf("nil return"), map[string]interface{}{
		"redis-url": "localhost:1111",
		"name":      "errorx",
	})
}
func tmpWrap() error {
	return NewFromString("nil return")
}
