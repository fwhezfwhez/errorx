package errorCollection

import (
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"log"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

var ec *ErrorCollection

func Init() {
	ec = NewCollection()
	ec.Add(errorx.NewFromString("an error happens"))
	ec.Add(errorx.NewFromString("another error happens"))
	//ec.AddHandler(LogEr())
	ec.HandleChain()
}

// Test handle()
func TestErrorCollection_Handle(t *testing.T) {
	Init()
	//prepare 2 errors
	runtime.Gosched()
	time.Sleep(3 * time.Second)
	// add an error when routine on
	ec.Add(errorx.NewFromString("after 3s, happen an error"))
	time.Sleep(1 * time.Second)
	ec.CloseHandles()

	time.Sleep(10 * time.Second)

	ec.HandleChain()
	time.Sleep(1 * time.Second)
	ec.Add(errorx.NewFromString("restart error"))
}

func TestHandle(t *testing.T) {
	Init()
	time.Sleep(2 * time.Second)
	wg := sync.WaitGroup{}
	wg.Add(10)
	for i := 0; i < 10; i++ {
		go func(a int, group *sync.WaitGroup) {
			ec.Add(errorx.NewFromString(strconv.Itoa(a) + ":error"))
			wg.Done()
		}(i, &wg)
	}
	wg.Wait()
	time.Sleep(10 * time.Second)
	fmt.Println("done")
	ec.CloseHandles()

}

// Test handle error by self design log
func LogEr() func(e error) {
	return func(e error) {
		log.SetFlags(log.Llongfile | log.LstdFlags)
		log.Println(e.Error())
	}
}

// Test Logger handler
func TestLogger(t *testing.T) {
	Init()
	//ec.Handle(Logger())
	ec.AddHandler(Panic())
	time.Sleep(3 * time.Second)
	ec.Add(errorx.NewFromString("after 3s, happen an error"))
	time.Sleep(1 * time.Second)
	ec.CloseHandles()
	time.Sleep(10 * time.Second)
}

// Test handleChain
func TestHandlerChain(t *testing.T) {
	ec := NewCollection()
	ec.Add(errorx.NewFromString("an error happens"))
	ec.Add(errorx.NewFromString("another error happens"))

	sendEmail := func(e error) {
		fmt.Println("send email:", e.Error())
	}
	ec.AddHandler(Logger(), Panic(), sendEmail)
	ec.HandleChain()

	time.Sleep(5 * time.Second)
	ec.Add(errorx.NewFromString("after 5s,an error occured"))
	time.Sleep(2 * time.Second)
	ec.CloseHandles()
	time.Sleep(4 * time.Second)
}

// test handlers with context and basic together
func TestHandlerChain2(t *testing.T) {
	ec := NewCollection()
	// with context
	ignoreError := func(e error, ctx *Context) bool {
		if strings.Contains(e.Error(), "an ignorable error happens") {
			return false
		}
		return true
	}

	errorPutContext := func(e error, ctx *Context) bool {
		ctx.Set("has-error", true)
		return true
	}

	errorGetContext := func(e error, ctx *Context) bool {
		if ctx.GetBool("has-error") {
			fmt.Println("found a error in according 'has-error'")
		}
		return true
	}

	ec.AddHandlerWithContext(ignoreError, errorPutContext, errorGetContext)

	ec.HandleChain()

	time.Sleep(5 * time.Second)
	ec.Add(errorx.NewFromString("after 5s,an error occured"))
	time.Sleep(2 * time.Second)

	ec.Add(errorx.NewFromStringf("an ignorable error happens"))
	ec.CloseHandles()
	time.Sleep(4 * time.Second)
}
