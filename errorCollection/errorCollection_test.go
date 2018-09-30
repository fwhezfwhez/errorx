package errorCollection

import (
	"fmt"
	"github.com/fwhezfwhez/errorx"
	"log"
	"runtime"
	"strconv"
	"sync"
	"testing"
	"time"
)
var ec *ErrorCollection
func Init(){
	ec  = NewCollection()
	ec.Add(errorx.NewFromString("an error happens"))
	ec.Add(errorx.NewFromString("another error happens"))
	ec.AddHandler(LogEr())
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
}

func TestHandle(t *testing.T){
	Init()
	time.Sleep(2*time.Second)
	wg := sync.WaitGroup{}
	wg.Add(10)
	for i:=0;i<10;i++{
		go func(a int,group *sync.WaitGroup){
			ec.Add(errorx.NewFromString(strconv.Itoa(a)+":error"))
			wg.Done()
		}(i,&wg)
	}
	wg.Wait()
	time.Sleep(10*time.Second)
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
