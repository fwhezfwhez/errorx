package main

import (
	"errors"
	"fmt"
	"github.com/fwhezfwhez/errorx"
)

func main() {
	if e := Control(); e != nil {
		e.(errorx.Error).PrintStackTrace()
		// log.Println(e.(errorx.Error).StackTrace())
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