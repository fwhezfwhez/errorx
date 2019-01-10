# errorx
[![Godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/fwhezfwhez/errorx)

[![Gitter](https://badges.gitter.im/Join Chat.svg)](https://gitter.im/fwhezfwhez-errorx/community)

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

		/*
		    		switch v := e.(type) {
            		case errorx.Error:
            			log.Println(strings.Join(v.StackTraces, "\n"))
            		case error:
            			log.Println(e.Error())
		*/
	}
}
func Email() func(e error) {
	return func(e error) {
		fmt.Println("sending an email,error:",e.Error())

				/*
        		    		switch v := e.(type) {
                    		case errorx.Error:
                    			fmt.Println(strings.Join(v.StackTraces, "\n"))
                    		case error:
		                        fmt.Println("sending an email,error:",e.Error())
        		*/
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

## Advantages
### 1.**No need to log everywhere!**
assume a project like:
```
main.go
control
  |__user_control.go
  |__order_control.go
  |__other_control.go
service
  |__user_service.go
  |__order_service.go
  |__other_service.go
dao
  |__user_dao.go
  |__order_dao.go
  |__other_dao.go
```

we used to handle it like:
```go
// order_dao.go
if er:= db.SQL("select * from order").Find(&orders);er!=nil{
    log.Println(er.Error)
    return er
}

```
```go
// order_service.go
if er:= orderDao.GetAll();er!=nil{
    log.Println(er.Error)
    return er
}
```
**.............**

This way has some fatal disadvantages:

1.**log.Println() will cause data escape which is bad for gc.And all log should be limitly used in product mode**.

2.**A same error instance has been log several times which is badly read**.

3.**If one or more error that may contain the same error.Error(),however log.Println is limited,It's hard to location the error spot.**

4.**To improve '3',you may add 'debug.Stack()' to bind with the error.Don't! because this will read the whole Stack without specific depth.It causes cpu and io busy.**

To improve above:
```go
// order_dao.go
if er:= db.SQL("select * from order").Find(&orders);er!=nil{
    return errorx.New(er)
}

```
```go
// order_service.go
if er:= orderDao.GetAll();er!=nil{
    return errorx.Wrap(er)
}
```
```go
// order_control.go
if er:= orderService.GetAll();er!=nil{
    handle(er.(errorx.Error).StackTrace())
    w.Write([]byte(er.Error()))
}
```

#### Why this improves?

reply 1:**er.(errorx.Error).StackTrace() is like**
```
G:/go_workspace/GOPATH/src/errorX/example/main.go: 26 | connect to mysql time out
G:/go_workspace/GOPATH/src/errorX/example/main.go: 34 | connect to mysql time out
G:/go_workspace/GOPATH/src/errorX/example/main.go: 42 | inner service error,please
```
**no need to log everywhere but log once**.

reply 2: **errorx only handle a same error**.

reply 3: **if two error.Error() looks the same like 'connect to mysql time out', it differs in spot path**.

reply 4: **stacktrace was recorded when error happen and pull a depth of 1 of runtime.Caller(-1).**

### **2. use errorChain and errox together**
assume a project like:
```
main.go
util
  |__error_garbage
          |__init.go
control
  |__user_control.go
  |__order_control.go
  |__other_control.go
service
  |__user_service.go
  |__order_service.go
  |__other_service.go
dao
  |__user_dao.go
  |__order_dao.go
  |__other_dao.go
```
util/error_garbage/init.go
```go
var Garbage *errorCollection.ErrorCollection

func init() {
    Garbage = errorCollection.NewCollection()
	Garbage.AddHandler(errorCollection.Fmt(), Sentry())
	Garbage.HandleChain()
}
// an example of handler
func Sentry() func(e error) {
	return func(e error) {
		fmt.Println("an error has been sent to Sentry")
		switch v := e.(type) {
		case errorx.Error:
			go raven.CaptureMessage("\n"+strings.Join(v.StackTraces, "\n"), map[string]string{"stackTrace": "\n" + strings.Join(v.StackTraces, "\n")})
		case error:
			go raven.CaptureMessage(v.Error(), map[string]string{"msg": v.Error()})
		}
	}
}
```
```go
// order_dao.go
if er:= db.SQL("select * from order").Find(&orders);er!=nil{
	return errorx.New(er)
}

```
```go
// order_service.go
if er:= orderDao.GetAll();er!=nil{
	return errorx.Wrap(er)
}
```
```go
// order_control.go
if er:= orderService.GetAll();er!=nil{
    // don't forget to import 'error_garbage' and rename as errorGarbage
    errorGarbage.Garbage.Add(er)
	w.Write([]byte(er.Error()))
}
```

