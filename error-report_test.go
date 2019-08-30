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
	rp.AddURL("dev", "http://localhost:8196").AddURL("pro", "http://localhost:8197")
	rp.AddModeHandler("dev", DefaultHandler).AddModeHandler("pro", rp.Mode("pro").ReportURLHandler)

	//_ = rp.Mode("dev").SaveError(errors.New("nil return"), map[string]interface{}{
	//	"api": "/xxx/yyy/",
	//})

	go rp.Mode("pro").SaveError(errors.New("nil return"), map[string]interface{}{
		"api": "/xxx/yyy/",
	})
	time.Sleep(10 * time.Second)
}
