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

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

## 1. What is different from the officials
| property | info | example | error | errorx |
|:----------- | :---- |:------|:-------------:|--:|
| Error() | what the error really is  | password wrong | yes | yes |
| StackTrace() | where the error went through| file.go :32 password wrong| no | yes |
| ReGen() | covering the real cause and return new error|  real_cause: time out -> user_view:inner service error | no | yes|
| errorChain | an error can be handled by handlers | | no | yes|

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
	rp = errorx.NewReporter()
	rp.AddModeHandler("dev", errorx.DefaultHandler)
}
func main() {
	e := fmt.Errorf("nil return")
	if e != nil {
		rp.Mode("dev").SaveError(errorx.Wrap(e), map[string]interface{}{
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
    "message": "2019-08-30 18:02:53 | G:/go_workspace/GOPATH/src/errorX/error-report.go: 24 | 2019-08-30 18:02:53 | G:/go_workspace/GOPATH/src/errorX/error-report.go: 134 | 2019-08-30 18:02:53 | G:/go_workspace/GOPATH/src/test_X/tmp/main.go: 17 | nil return",
    "age": 1,
    "error_uuid": "0ad565cd-7fb3-4ee9-bbc7-db3abcd0aebb",
    "username": "errorx"
}
```

**Using url report**

error will POST into url

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
	rp = errorx.NewReporter()
	rp.AddURL("pro", "http://localhost:9191")
	rp.AddModeHandler("pro", rp.Mode("pro").ReportURLHandler)
}
func main() {
	e := fmt.Errorf("nil return")
	if e != nil {
		_ = rp.Mode("pro").SaveError(errorx.Wrap(e), map[string]interface{}{
			"username": "errorx",
			"age":      1,
		})
		return
	}
}

```
Output in server panel:
```
Recv: {
    "age": 1,
    "error_uuid": "1fd36ad7-9529-441d-97f8-2f6cc7886aae",
    "message": "2019-08-30 18:15:37 | G:/go_workspace/GOPATH/src/errorX/error-report.go: 39 | 2019-08-30 18:15:37 | G:/go_workspace/GOPATH/src/errorX/error-report.go: 133 | 2019-08-30 18:15:37 | G:/go_workspace/GOPATH/src/test_X/tmp/main.go: 18 | nil return\n\n\n",
    "username": "errorx"
  }
```
