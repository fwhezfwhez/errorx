# errorx
[![Godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/fwhezfwhez/errorx)

a very convenient error handler.
## What is different from the officials
| property | info | example | error | errorx |
|:----------- | :---- |:------|:-------------:|--:|
| Error() | what the error really is  | password wrong | yes | yes |
| StackTrace() | where the error went through| file.go :32 password wrong| no | yes |
| ReGen() | covering the real cause and return new error|  real_cause: time out -> user_view:inner service error | no | yes|
| errorChain | an error can be handled by handlers | | no | yes|

## start
`go get github.com/fwhezfwhez/errorx`

## Basic
```go
package main

import (
	"github.com/fwhezfwhez/errorx"
	"errors"
	"fmt"
)

func main() {
	if e := Control(); e != nil {
		e.(errorx.Error).PrintStackTrace()
		//log.Println(e.(errorx.Error).StackTrace())
	} else {
		Reply()
	}
}

// assume an engine to connect mysql
func DB() error {
	return errors.New("connect to mysql time out")
}

// handle database operation
func Dao() error {
	if er := DB(); er != nil {
		return errorx.New(er)
	}
	return nil
}

// handle logic service
func Service() error {
	if er := Dao(); er != nil {
		return errorx.Wrap(er)
	}
	return nil
}

// handle request distribute from main
func Control() error {
	if er := Service(); er != nil {
		return errorx.ReGen(er, errors.New("inner service error,please call admin for help"))
	}
	return nil
}

// reply a the request
func Reply(){
	fmt.Println("handle success")
}
```

result:
```
StackTrace | CausedBy
G:/go_workspace/GOPATH/src/errorX/example/main.go: 26 | connect to mysql time out
G:/go_workspace/GOPATH/src/errorX/example/main.go: 34 | connect to mysql time out
G:/go_workspace/GOPATH/src/errorX/example/main.go: 42 | inner service error,please call admin for help

```

## ErrorChain
```go
package main

import (
	"errors"
	"fmt"
	"github.com/fwhezfwhez/errorx/errorCollection"
	"log"
	"time"
)

func main() {
	fmt.Println("test error chain handlers")
	// init
	ec := errorCollection.NewCollection()
	// add handler
	ec.AddHandler(LogEr(), Email(), ReportToCloud(), HandleAsNewRoutine())
	// begin handling and wait for errors
	ec.HandleChain()

	// assume an error occur
	ec.Add(errors.New("I am an error"))

	time.Sleep(2*time.Second)
	// an error happen after 2 seconds
	ec.Add(errors.New("I occur after 2s"))
	// make sure the handling has its running time
	time.Sleep(10* time.Second)
}
func LogEr() func(e error) {
	return func(e error) {
		log.SetFlags(log.Llongfile | log.LstdFlags)
		log.Println(e.Error())
	}
}
func Email() func(e error) {
	return func(e error) {
		fmt.Println("sending an email,error:",e.Error())
	}
}

func ReportToCloud() func(e error) {
	return func(e error) {
		fmt.Println("reporting the error to cloud,error:", e.Error())
	}
}

func HandleAsNewRoutine() func(e error) {
	return func(e error) {
		go func() {
			fmt.Println("handler the error routinely in case that the handler would cost much time")
		}()
	}
}

```