package errorx

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func startReportServer(port string) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		buf, e := ioutil.ReadAll(r.Body)
		if e != nil {
			fmt.Println(Wrap(e).Error())
			return
		}
		fmt.Println("Recv:", string(buf))
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})
	server := &http.Server{
		Addr:         port,
		WriteTimeout: time.Second * 3, //设置3秒的写超时
		Handler:      mux,
	}
	server.ListenAndServe()
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
	rp := NewReporter("pro")
	rp.SetContextName("request")
	rp.AddURL("dev", "http://localhost:8196").
		AddURL("pro", "http://localhost:8197")
	rp.AddModeHandler("dev", rp.DefaultHandler).
		AddModeHandler("pro", rp.ReportURLHandler)

	_ = rp.Mode("dev").SaveError(errors.New("nil return"), map[string]interface{}{
		"api": "/xxx/yyy/",
	})
	_ = rp.Mode("dev").SaveError(errors.New("nil return"), nil)

	go rp.SaveError(errors.New("nil return"), map[string]interface{}{
		"api": "/xxx/yyy/",
	})
	go rp.SaveError(errors.New("nil return"), nil)
	time.Sleep(10 * time.Second)
}

func TestReporter_JSONIndent_JSON(t *testing.T) {
	rp := NewReporter("pro")

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

	eUuid, buf, e = JSON(NewFromString("nil return"), map[string]interface{}{
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

	eUuid, buf, e = JSONIndent(NewFromString("nil return"), map[string]interface{}{
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

	rp := NewReporter("pro")
	rp.AddURL("dev", "http://localhost:8196").
		AddURL("pro", "http://localhost:7123")
	rp.AddModeHandler("dev", DefaultHandler).
		AddModeHandler("pro", rp.Mode("pro").ReportURLHandler)

	<-serverStart
	for i := 0; i < 100; i++ {
		go rp.Mode("dev").SaveError(errors.New("nil return"), map[string]interface{}{
			"api": "/xxx/yyy/",
		})

		go rp.Mode("pro").SaveError(errors.New("nil return"), map[string]interface{}{
			"api": "/xxx/yyy/",
		})
	}
	time.Sleep(20 * time.Second)
}
