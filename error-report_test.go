package errorx

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"testing"
	"time"
)

func startReportServer(port string) {
	r := gin.Default()

	r.POST("/", func(c *gin.Context) {
		buf, e := ioutil.ReadAll(c.Request.Body)
		if e != nil {
			fmt.Println(Wrap(e).Error())
			return
		}
		fmt.Println("Recv:", string(buf))
	})
	r.Run(port)
}

func TestReporter(t *testing.T) {
	var serverStart = make(chan int, 0)

	go func() {
		go startReportServer(":8196")
		go startReportServer(":8197")
		time.Sleep(2 * time.Second)
		serverStart <- 1
	}()

	<-serverStart
	rp := NewReporter()
	rp.AddURL("dev", "http://localhost:8196").
		AddURL("pro", "http://localhost:8197")
	rp.AddModeHandler("dev", DefaultHandler).
		AddModeHandler("pro", rp.Mode("pro").ReportURLHandler)

	_ = rp.Mode("dev").SaveError(errors.New("nil return"), map[string]interface{}{
		"api": "/xxx/yyy/",
	})
	_ = rp.Mode("dev").SaveError(errors.New("nil return"), nil)

	go rp.Mode("pro").SaveError(errors.New("nil return"), map[string]interface{}{
		"api": "/xxx/yyy/",
	})
	go rp.Mode("pro").SaveError(errors.New("nil return"), nil)
	time.Sleep(10 * time.Second)
}

func TestReporter_JSONIndent_JSON(t *testing.T) {
	rp := NewReporter()

	eUuid, buf, e := rp.JSON(NewFromString("nil return"), map[string]interface{}{
		"api": "/xx/xxx/xx",
	})
	if e != nil {
		fmt.Println(e.Error())
		t.Fail()
		return
	}
	fmt.Println(eUuid)
	fmt.Println(string(buf))

	eUuid, buf, e = rp.JSONIndent(NewFromString("nil return"), map[string]interface{}{
		"api": "/xx/xxx/xx",
	}, "  ", "  ")
	if e != nil {
		fmt.Println(e.Error())
		t.Fail()
		return
	}
	fmt.Println(eUuid)
	fmt.Println(string(buf))
}

func TestNewReporterConcurrent(t *testing.T) {
	var serverStart = make(chan int, 0)
	go func() {
		go startReportServer(":7123")
		time.Sleep(2 * time.Second)
		serverStart <- 1
	}()

	rp := NewReporter()
	rp.AddURL("dev", "http://localhost:8196").
		AddURL("pro", "http://localhost:7123")
	rp.AddModeHandler("dev", DefaultHandler).
		AddModeHandler("pro", rp.Mode("pro").ReportURLHandler)

	<-serverStart
	for i:=0;i<100;i++ {
		go rp.Mode("dev").SaveError(errors.New("nil return"), map[string]interface{}{
			"api": "/xxx/yyy/",
		})

		go rp.Mode("pro").SaveError(errors.New("nil return"), map[string]interface{}{
			"api": "/xxx/yyy/",
		})
	}
	time.Sleep(20 * time.Second)
}
