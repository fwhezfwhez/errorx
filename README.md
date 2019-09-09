# errorx
[![Godoc](http://img.shields.io/badge/godoc-reference-blue.svg?style=flat)](https://godoc.org/github.com/fwhezfwhez/errorx)
[![Build Status]( https://www.travis-ci.org/fwhezfwhez/errorx.svg?branch=master)]( https://www.travis-ci.org/fwhezfwhez/errorx)
[![Gitter](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/fwhezfwhez-errorx/community)

a very convenient error handler.

<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [1. What is different from the officials](#1-what-is-different-from-the-officials)
- [2. Start](#2-start)
- [3. Module](#3-module)
    - [3.1 Error stacktrace](#31-error-stacktrace)
    - [3.2 Error Report](#32-error-report)
    - [3.3 JSON](#33-json)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## 1. What is different from the officials

- Supporting generate error stacktrace, locating more efficiently than runtime stacktrace.
- Supporting report error to HTTP URL.
- Supporting generate json-marshalled error detail for outer pipeline.
- Supporting differrent mode uses different error handlers.

## 2. Start
`go get github.com/fwhezfwhez/errorx`

## 3. Module

#### 3.1 Error stacktrace
```go
package main

import (
	"fmt"
	"github.com/fwhezfwhez/errorx"
)

func main() {
	e := fmt.Errorf("nil return")
	fmt.Println(errorx.Wrap(e).Error())
}

```

Output

```go
2019-08-30 17:51:42 | G:/go_workspace/GOPATH/src/test_X/tmp/main.go: 10 | nil return
```

#### 3.2 Error Report

**Using defaultHandler print in console**

error will be log as json on console.
```go
package main

import (
	"github.com/fwhezfwhez/errorx"
	"fmt"
)

var rp *errorx.Reporter

func init() {
	rp = errorx.NewReporter("dev")
	rp.AddModeHandler("dev", rp.DefaultHandler)
}
func main() {
	e := fmt.Errorf("nil return")
	if e != nil {
		rp.SaveError(errorx.Wrap(e), map[string]interface{}{
			"username": "errorx",
			"age":      1,
		})
		return
	}
}
```
output:
```go
{
    "context": {
      "api": "/xxx/yyy/"
    },
    "error_uuid": "11d35e60-5abc-462d-9df1-bb5b01d79807",
    "message": "2019-08-31 08:58:29 | G:/go_workspace/GOPATH/src/errorX/error-report.go: 123 | err 'nil return' \n goroutine 51 [running]:\nruntime/debug.Stack(0xc0002022e0, 0xb3604d, 0xa)\n\tE:/go1.12/src/runtime/debug/stack.go:24 +0xa4\nerrorX.Reporter.SaveError(0xc0001fc1e0, 0xb30f5d, 0x3, 0xc0001fc180, 0x0, 0x0, 0x0, 0xc0001fc1b0, 0x0, 0x0, ...)\n\tG:/go_workspace/GOPATH/src/errorX/error-report.go:123 +0x30b\ncreated by errorX.TestReporter\n\tG:/go_workspace/GOPATH/src/errorX/error-report_test.go:45 +0x702\n\n2019-08-31 08:58:29 | G:/go_workspace/GOPATH/src/errorX/error-report.go: 144 | err 'nil return' \n goroutine 51 [running]:\nruntime/debug.Stack(0xc0002022e0, 0xb3604d, 0xa)\n\tE:/go1.12/src/runtime/debug/stack.go:24 +0xa4\nerrorX.Reporter.SaveError(0xc0001fc1e0, 0xb30f5d, 0x3, 0xc0001fc180, 0x0, 0x0, 0x0, 0xc0001fc1b0, 0x0, 0x0, ...)\n\tG:/go_workspace/GOPATH/src/errorX/error-report.go:123 +0x30b\ncreated by errorX.TestReporter\n\tG:/go_workspace/GOPATH/src/errorX/error-report_test.go:45 +0x702\n\n2019-08-31 08:58:29 | G:/go_workspace/GOPATH/src/errorX/error-report.go: 49 | err 'nil return' \n goroutine 51 [running]:\nruntime/debug.Stack(0xc0002022e0, 0xb3604d, 0xa)\n\tE:/go1.12/src/runtime/debug/stack.go:24 +0xa4\nerrorX.Reporter.SaveError(0xc0001fc1e0, 0xb30f5d, 0x3, 0xc0001fc180, 0x0, 0x0, 0x0, 0xc0001fc1b0, 0x0, 0x0, ...)\n\tG:/go_workspace/GOPATH/src/errorX/error-report.go:123 +0x30b\ncreated by errorX.TestReporter\n\tG:/go_workspace/GOPATH/src/errorX/error-report_test.go:45 +0x702\n\n"
}
```

**Using url report**

error will POST into specific url.

server
```go
package main

import (
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"github.com/gin-gonic/gin"
	"io/ioutil"
)

func main() {

	r := gin.Default()

	r.POST("/", func(c *gin.Context) {
		buf, e := ioutil.ReadAll(c.Request.Body)
		if e != nil {
			fmt.Println(errorx.Wrap(e).Error())
			return
		}
		fmt.Println("Recv:", string(buf))
	})
	r.Run(":9191")
}

```

```go
package main

import (
	"github.com/fwhezfwhez/errorx"
	"fmt"
)

var rp *errorx.Reporter

func init() {
	rp = errorx.NewReporter("pro")
	rp.AddURL("pro", "http://localhost:9191")
	rp.AddURL("dev", "http://localhost:9192")
	rp.AddModeHandler("pro", rp.ReportURLHandler)
	rp.AddModeHandler("dev", rp.Mode("dev").DefaultHandler)
}
func main() {
	e := fmt.Errorf("nil return")
	if e != nil {
	    // rp's mode is pro, it will send error to localhost:9191
		_ = rp.SaveError(errorx.Wrap(e), map[string]interface{}{
			"username": "errorx",
			"age":      1,
		})
		// clone a rp and reset its mode to dev, it will print error in console by DefaultHandler
        _ = rp.Mode("dev").SaveError(errorx.Wrap(e), map[string]interface{}{
            "username": "errorx",
            "age":      1,
        })
		return
	}
}

```
Output in server panel:
```
Recv:
{
    "context": {
      "api": "/xxx/yyy/"
    },
    "error_uuid": "11d35e60-5abc-462d-9df1-bb5b01d79807",
    "message": "2019-08-31 08:58:29 | G:/go_workspace/GOPATH/src/errorX/error-report.go: 123 | err 'nil return' \n goroutine 51 [running]:\nruntime/debug.Stack(0xc0002022e0, 0xb3604d, 0xa)\n\tE:/go1.12/src/runtime/debug/stack.go:24 +0xa4\nerrorX.Reporter.SaveError(0xc0001fc1e0, 0xb30f5d, 0x3, 0xc0001fc180, 0x0, 0x0, 0x0, 0xc0001fc1b0, 0x0, 0x0, ...)\n\tG:/go_workspace/GOPATH/src/errorX/error-report.go:123 +0x30b\ncreated by errorX.TestReporter\n\tG:/go_workspace/GOPATH/src/errorX/error-report_test.go:45 +0x702\n\n2019-08-31 08:58:29 | G:/go_workspace/GOPATH/src/errorX/error-report.go: 144 | err 'nil return' \n goroutine 51 [running]:\nruntime/debug.Stack(0xc0002022e0, 0xb3604d, 0xa)\n\tE:/go1.12/src/runtime/debug/stack.go:24 +0xa4\nerrorX.Reporter.SaveError(0xc0001fc1e0, 0xb30f5d, 0x3, 0xc0001fc180, 0x0, 0x0, 0x0, 0xc0001fc1b0, 0x0, 0x0, ...)\n\tG:/go_workspace/GOPATH/src/errorX/error-report.go:123 +0x30b\ncreated by errorX.TestReporter\n\tG:/go_workspace/GOPATH/src/errorX/error-report_test.go:45 +0x702\n\n2019-08-31 08:58:29 | G:/go_workspace/GOPATH/src/errorX/error-report.go: 49 | err 'nil return' \n goroutine 51 [running]:\nruntime/debug.Stack(0xc0002022e0, 0xb3604d, 0xa)\n\tE:/go1.12/src/runtime/debug/stack.go:24 +0xa4\nerrorX.Reporter.SaveError(0xc0001fc1e0, 0xb30f5d, 0x3, 0xc0001fc180, 0x0, 0x0, 0x0, 0xc0001fc1b0, 0x0, 0x0, ...)\n\tG:/go_workspace/GOPATH/src/errorX/error-report.go:123 +0x30b\ncreated by errorX.TestReporter\n\tG:/go_workspace/GOPATH/src/errorX/error-report_test.go:45 +0x702\n\n"
}
```

#### 3.3 JSON

JSON and JSONIndent will generate a json buf from error and context.

```go
package main

import (
	"github.com/fwhezfwhez/errorx"
	"fmt"
)

func main() {
	eUuid, buf, e := errorx.JSON(errorx.NewFromString("nil return"), map[string]interface{}{
		"api": "/xx/xxx/xx",
	})
	fmt.Println(eUuid)
	fmt.Println(string(buf))
	fmt.Println(e)
}
```
Output:
```
befe4742-6905-4817-96aa-19fdb63bd83f
{"context":{"api":"/xx/xxx/xx"},"error_uuid":"befe4742-6905-4817-96aa-19fdb63bd83f","message":"2019-08-31 09:18:41 | G:/go_workspace/GOPATH/src/errorX/error-report_test.go: 56 | nil return\n2019-08-31 09:18:41 | G:/go_workspace/GOPATH/src/errorX/error-report.go: 171 | nil return\n"}
```
