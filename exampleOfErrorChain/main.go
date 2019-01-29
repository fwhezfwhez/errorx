package main

import (
	"errors"
	"fmt"
	"github.com/fwhezfwhez/errorx"
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

	time.Sleep(2 * time.Second)
	// an error happen after 2 seconds
	ec.Add(errorx.NewFromString("I occur after 2s"))

	fmt.Printf("test add nil")
	// make sure the handling has its running time
	time.Sleep(10 * time.Second)
}
func LogEr() func(e error) {
	return func(e error) {
		log.SetFlags(log.Llongfile | log.LstdFlags)
		log.Println(e.Error())
	}
}
func Email() func(e error) {
	return func(e error) {
		fmt.Println("sending an email,error:", e.Error())
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
			/*
			switch v := e.(type) {
			case errorx.Error:
				fmt.Println(strings.Join(v.StackTraces, "\n"))
			case error:
				fmt.Println(e.Error())
           */
			fmt.Println("handler the error routinely in case that the handler would cost much time")
		}()
	}
}
