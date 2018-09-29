package errorCollection

import (
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"log"
	"testing"
	"time"
)

// Test handle()
func TestErrorCollection_Handle(t *testing.T) {
	ec := NewCollection()

	//prepare 2 errors
	ec.Add(errorx.NewFromString("an error happens"))
	ec.Add(errorx.NewFromString("another error happens"))
	ec.Handle(LogEr())
	time.Sleep(3 * time.Second)
	// add an error when routine on
	ec.Add(errorx.NewFromString("after 3s, happen an error"))
	time.Sleep(1 * time.Second)
	ec.CloseHandles()

	time.Sleep(10 * time.Second)
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
	ec := NewCollection()
	ec.Add(errorx.NewFromString("an error happens"))
	ec.Add(errorx.NewFromString("another error happens"))

	//ec.Handle(Logger())
	ec.Handle(Panic())
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
