package errorCollection

import (
	"fmt"
	"github.com/fwhezfwhez/errorx"
	queue "github.com/fwhezfwhez/go-queue"
	"log"
	"strings"
	"sync"
	"time"
)

type ErrorBox *queue.Queue
type ErrorHandler func(e error)
type ErrorHandlerWithContextInSeries func(e error, ctx *Context) (next bool)

type ErrorCollection struct {
	errors                              ErrorBox                          // store errors in queue
	ErrorHandleChain                    []ErrorHandler                    // handler to deal with error
	ErrorHandleChainWithContextInSeires []ErrorHandlerWithContextInSeries // handler with a context and a flag 'next' to decide whether to handle next handler
	M                                   *sync.Mutex                       // field lock RWlock
	CatchErrorChan                      chan error                        // when error in queue,it will be put into catchErrorChan
	AutoHandleChan                      chan int                          // used to close the channel
}

// New a error collection
func NewCollection() *ErrorCollection {
	log.SetFlags(log.Llongfile | log.LstdFlags)
	return &ErrorCollection{
		errors:                              queue.NewCap(200),
		ErrorHandleChain:                    make([]ErrorHandler, 0, 10),
		ErrorHandleChainWithContextInSeires: make([]ErrorHandlerWithContextInSeries, 0, 10),
		CatchErrorChan:                      make(chan error, 1),
		AutoHandleChan:                      make(chan int, 1),
		M:                                   &sync.Mutex{},
	}
}
func Default() *ErrorCollection {
	e := NewCollection()
	e.AddHandler(Logger())
	return e
}
func (ec *ErrorCollection) GetQueueLock() *sync.Mutex {
	return ec.errors.Mutex
}

// Return its error number
func (ec *ErrorCollection) SafeLength() int {
	return (*queue.Queue)(ec.errors).SafeLength()
}

// Return its error number
func (ec *ErrorCollection) Length() int {
	return (*queue.Queue)(ec.errors).Length()
}

// Add an error into collect
func (ec *ErrorCollection) Add(e error) {
	ec.M.Lock()
	defer ec.M.Unlock()

	if len(ec.CatchErrorChan) < cap(ec.CatchErrorChan) {
		ec.CatchErrorChan <- e
	} else {
		(*queue.Queue)(ec.errors).Push(e)
	}
	defer func() {
		if r := recover(); r != nil {
			log.Printf("catch a panic：%s\n", r)
		}
	}()
}


// This is a self design  function to handle the inner errors collected via a single handler typed 'ErrorHandler'
func (ec *ErrorCollection) Handle(f ErrorHandler) {
	ec.M.Lock()
	ec.newAutoHandleChan()
	ec.CatchError()
	ec.M.Unlock()
	go func() {
		log.Println("handle routine starts,use ec.CloseHandles to stop")
	L:
		for {
			select {
			case <-ec.AutoHandleChan:
				log.Println("handle finish by auto handle chan")
				break L
			case e := <-ec.CatchErrorChan:
				f(e)
			}
		}
	}()
}

// this is a self design function to handle the inner errors collected via a single handler typed 'ErrorHandlerWithContextInSeries'
// DO not Use handler in series typed 'ErrorHandlerWithContextInSeries' together with basic typed 'ErrorHandler', because they share the error channel causing a single error
// could be handler only by one the them.
func (ec *ErrorCollection) HandleInSeries(f ErrorHandlerWithContextInSeries) {
	ec.M.Lock()
	ec.newAutoHandleChan()
	ec.CatchError()
	ec.M.Unlock()
	go func() {
		log.Println("handle routine starts,use ec.CloseHandles to stop")
	L:
		for {
			select {
			case <-ec.AutoHandleChan:
				log.Println("handle finish by auto handle chan")
				break L
			case e := <-ec.CatchErrorChan:
				// next
				ctx := NewContext()
				_ = f(e, ctx)
			}
		}
	}()
}

func (ec *ErrorCollection) safeNewAutoHandleChan() {
	ec.M.Lock()
	defer ec.M.Unlock()
	ec.AutoHandleChan = make(chan int, 1)
}
func (ec *ErrorCollection) newAutoHandleChan() {
	ec.AutoHandleChan = make(chan int, 1)
}

func (ec *ErrorCollection) CloseHandles() {
	ec.M.Lock()
	defer ec.M.Unlock()
	close(ec.AutoHandleChan)
}

func (ec *ErrorCollection) IfErrorChanFull() bool {
	if len(ec.CatchErrorChan) < cap(ec.CatchErrorChan) {
		return false
	}
	return true
}

// Handle the error queue one by one  by those handler added
// How to add a handler>
// ec.AddHandler(Logger(),Panic(),SendEmail()) ...
// How to make handler routine dependent?
// ** When should do like this?**
// ** When you realize the handler might risk timing out or**
//	func LogEr() func(e error) {
//		return func(e error) {
//			go func(er error) {
//				log.SetFlags(log.Llongfile | log.LstdFlags)
//				log.Println(e.Error())
//          }(e)
//		}
//	}
//
func (ec *ErrorCollection) HandleChain() {
	ec.newAutoHandleChan()
	ec.CatchError()
	go func() {
		log.Println("handleChain routine starts,use ec.CloseHandles to stop")
	L:
		for {
			select {
			case e := <-ec.CatchErrorChan:
				// handle basic handle chains
				for _, f := range ec.ErrorHandleChain {
					f(e)
				}

				// handle handle chains with a context and next control
				ctx := NewContext()
				for _, f := range ec.ErrorHandleChainWithContextInSeires {
					next := f(e, ctx)
					if !next {
						break L
					}
				}
			case <-ec.AutoHandleChan:
				log.Println("handleChain finish by auto handle chan")
				break L
			}
		}
	}()

}

// Add handler to handler chain
func (ec *ErrorCollection) AddHandler(handler ... ErrorHandler) {
	ec.M.Lock()
	defer ec.M.Unlock()
	ec.ErrorHandleChain = append(ec.ErrorHandleChain, handler...)
}

// Add handler with context to handler chain
func (ec *ErrorCollection) AddHandlerWithContext(handler ... ErrorHandlerWithContextInSeries) {
	ec.M.Lock()
	defer ec.M.Unlock()
	ec.ErrorHandleChainWithContextInSeires = append(ec.ErrorHandleChainWithContextInSeires, handler...)
}

// Clear errors in collection
func (ec *ErrorCollection) Clear() {
	ec.M.Lock()
	defer ec.M.Unlock()
	ec.errors = ErrorBox(queue.NewCap(500))
}

// Pop an error.
// The popped error is from queue' head.
// When use Pop(), it means the error has been dealed and deleted from the queue
func (ec *ErrorCollection) Pop() error {
	q := (*queue.Queue)(ec.errors)
	ec.GetQueueLock().Lock()
	defer ec.GetQueueLock().Unlock()
	//if ec.Length() > 0 {
	v := q.Pop()
	if v != nil {
		e := v.(error)
		return e
	}
	return nil
	//}
	return nil
}

// Get an error
// The error is from queue' head.
// When use Get(), it means the error is only for query,not deleted from the queue
func (ec *ErrorCollection) GetError() error {
	q := (*queue.Queue)(ec.errors)
	if ec.Length() != 0 {
		h, _ := q.SafeValidHead()
		return errorx.Wrap(h.(error))
	}
	return nil
}

// When an error in , it will be pop into a chanel waiting for handling
func (ec *ErrorCollection) CatchError() <-chan error {
	go func() {
	L:
		for {
			select {
			case <-ec.AutoHandleChan:
				log.Println("close error catching by auto handle chan")
				break L
			default:
				ec.M.Lock()
				if x := ec.Pop(); x != nil {
					ec.CatchErrorChan <- x
				}
				ec.M.Unlock()
			}
			time.Sleep(200 * time.Millisecond)
		}
	}()
	return ec.CatchErrorChan
}

// Logger records error msg in log without modifying error queue.
func Logger() func(e error) {
	return func(e error) {
		log.SetFlags(log.LstdFlags | log.Llongfile)
		log.Println(e)
	}
}

// Fmt
func Fmt() func(e error) {
	return func(e error) {
		switch v := e.(type) {
		case errorx.Error:
			fmt.Println(strings.Join(v.StackTraces, "\n"))
		default:
			fmt.Println(v.Error())
		}
	}
}

// Panic records the first error and recover.
// Panic aims to fix the panic cause error and ignore other errors unless you fix the error up
func Panic() func(e error) {
	return func(e error) {
		defer func() {
			if r := recover(); r != nil {
				log.SetFlags(log.LstdFlags | log.Llongfile)
				log.Println("panic and recover,because of:", r)
			}
		}()
		 panic(e.Error())
	}
}
