package errorx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime/debug"
	"sync"
	"time"

	"github.com/gofrs/uuid"
)

// It will fmt error log into console.
var DefaultHandler = func(e error, context map[string]interface{}) {
	var tmp = make(map[string]interface{}, 0)
	tmp["error_uuid"] = context["error_uuid"]
	delete(context, "error_uuid")
	tmp["message"] = Wrap(e).Error()
	tmp["context"] = context

	var buf = []byte("")
	var er error
	buf, er = json.MarshalIndent(tmp, "  ", "  ")
	if er != nil {
		buf = []byte(Wrap(er).Error())
	}
	fmt.Println(string(buf))
}

type Reporter struct {
	c *http.Client

	mode string
	Url  map[string]string
	l1   sync.RWMutex

	HandleMode map[string]func(e error, context map[string]interface{})
	l2         sync.RWMutex

	renameOfContext string
}

func (r *Reporter) SetContextName(name string) {
	r.renameOfContext = name
}
func (r *Reporter) SetMode(mode string) {
	r.mode = mode
}

// rp.Mode related, should call like rp.Mode("dev").ReportURLHandler
// It will post error with context to an url storaged in reporter.Url
func (r Reporter) ReportURLHandler(e error, context map[string]interface{}) {
	if r.mode == "" {
		context["reporter"] = NewFromStringf("reporter mode empty, please call er.Mode('pro') first").Error()
		DefaultHandler(NewFromStringf("reporter mode empty, please call er.Mode('pro') first"), context)
		return
	}

	var tmp = make(map[string]interface{}, 0)

	tmp["error_uuid"] = context["error_uuid"]
	delete(context, "error_uuid")

	tmp["message"] = Wrap(e).Error()

	if r.renameOfContext != "" {
		tmp[r.renameOfContext] = context
	} else {
		tmp["context"] = context
	}

	buf, er := json.MarshalIndent(tmp, "  ", "  ")
	if er != nil {
		context["reporter"] = er.Error()
		DefaultHandler(Wrap(e), context)
		return
	}
	req, er := http.NewRequest("POST", r.Url[r.mode], bytes.NewReader(buf))
	req.Header.Set("Content-Type", "application/json")
	if er != nil {
		context["reporter"] = er.Error()
		DefaultHandler(Wrap(e), context)
		return
	}
	resp, er := r.c.Do(req)
	if er != nil {
		context["reporter"] = er.Error()
		DefaultHandler(Wrap(e), context)
		return
	}
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
}

func (r Reporter) DefaultHandler(e error, context map[string]interface{}) {
	var tmp = make(map[string]interface{}, 0)
	tmp["error_uuid"] = context["error_uuid"]
	delete(context, "error_uuid")
	tmp["message"] = Wrap(e).Error()
	if r.renameOfContext != "" {
		tmp[r.renameOfContext] = context
	} else {
		tmp["context"] = context
	}

	var buf = []byte("")
	var er error
	buf, er = json.MarshalIndent(tmp, "  ", "  ")
	if er != nil {
		buf = []byte(Wrap(er).Error())
	}
	fmt.Println(string(buf))
}

func NewReporter(mode string) *Reporter {
	return &Reporter{
		c: &http.Client{Timeout: 15 * time.Second},

		mode:       mode,
		Url:        make(map[string]string, 0),
		HandleMode: make(map[string]func(e error, context map[string]interface{})),
		l1:         sync.RWMutex{},
		l2:         sync.RWMutex{},
	}
}

func (r *Reporter) copy() *Reporter {
	var clone = &Reporter{
		c: r.c,

		mode:       r.mode,
		Url:        r.Url,
		HandleMode: r.HandleMode,
		l1:         r.l1,
		l2:         r.l2,

		renameOfContext: r.renameOfContext,
	}
	return clone
}

// Mode will clone a copy of reporter to avoid different goroutine using a unsafe clone
func (r *Reporter) Mode(mode string) *Reporter {
	clone := r.copy()
	clone.mode = mode
	return clone
}

func (r *Reporter) AddURL(mode string, url string) *Reporter {
	r.Url[mode] = url
	return r
}
func (r *Reporter) AddModeHandler(mode string, f func(e error, context map[string]interface{})) *Reporter {
	r.HandleMode[mode] = f
	return r
}

// rp.Mode related, should call as r.Mode("dev").SaveError()
func (r Reporter) SaveError(e error, context map[string]interface{}) string {
L:
	switch v := e.(type) {
	case Error:
		break L
	case error:
		return r.SaveError(NewFromString(string(fmt.Sprintf("err '%s' \n %s", v.Error(), debug.Stack()))), context)
	}

	if context == nil {
		context = make(map[string]interface{}, 0)
	}

	var handler func(e error, context map[string]interface{})
	// judge whether exist handler for mode
	func() {
		var ok bool
		r.l2.RLock()
		defer r.l2.RUnlock()
		handler, ok = r.HandleMode[r.mode]
		if !ok {
			handler = DefaultHandler
		}
	}()
	u, _ := NewV4()
	errorUUID := u.String()
	context["error_uuid"] = errorUUID
	handler(Wrap(e), context)

	return errorUUID
}

// rp.Mode unrelated
// returns errorUUID, json-buf, error
func (r Reporter) JSON(e error, context map[string]interface{}) (string, []byte, error) {
L:
	switch v := e.(type) {
	case Error:
		break L
	case error:
		return r.JSON(NewFromString(string(fmt.Sprintf("err '%s' \n %s", v.Error(), debug.Stack()))), context)
	}

	if context == nil {
		context = make(map[string]interface{}, 0)
	}

	var tmp = make(map[string]interface{}, 0)

	u, _ := uuid.NewV4()
	errUUID := u.String()
	tmp["error_uuid"] = errUUID
	tmp["message"] = Wrap(e).Error()
	tmp["context"] = context
	buf, e := json.Marshal(tmp)
	if e != nil {
		return "", nil, Wrap(e)
	}
	return errUUID, buf, nil
}

func JSON(e error, context map[string]interface{}) (string, []byte, error) {
L:
	switch v := e.(type) {
	case Error:
		break L
	case error:
		return JSON(NewFromString(string(fmt.Sprintf("err '%s' \n %s", v.Error(), debug.Stack()))), context)
	}

	if context == nil {
		context = make(map[string]interface{}, 0)
	}

	var tmp = make(map[string]interface{}, 0)

	u, _ := uuid.NewV4()
	errUUID := u.String()
	tmp["error_uuid"] = errUUID
	tmp["message"] = Wrap(e).Error()
	tmp["context"] = context
	buf, e := json.Marshal(tmp)
	if e != nil {
		return "", nil, Wrap(e)
	}
	return errUUID, buf, nil
}

func (r Reporter) JSONIndent(e error, context map[string]interface{}, prefix, indent string) (string, []byte, error) {
L:
	switch v := e.(type) {
	case Error:
		break L
	case error:
		return r.JSON(NewFromString(string(fmt.Sprintf("err '%s' \n %s", v.Error(), debug.Stack()))), context)
	}

	if context == nil {
		context = make(map[string]interface{}, 0)
	}

	var tmp = make(map[string]interface{}, 0)

	u, _ := uuid.NewV4()
	errUUID := u.String()
	tmp["error_uuid"] = errUUID
	tmp["message"] = Wrap(e).Error()
	tmp["context"] = context
	buf, e := json.MarshalIndent(tmp, prefix, indent)
	if e != nil {
		return "", nil, Wrap(e)
	}
	return errUUID, buf, nil
}

func JSONIndent(e error, context map[string]interface{}, prefix, indent string) (string, []byte, error) {
L:
	switch v := e.(type) {
	case Error:
		break L
	case error:
		return JSON(NewFromString(string(fmt.Sprintf("err '%s' \n %s", v.Error(), debug.Stack()))), context)
	}

	if context == nil {
		context = make(map[string]interface{}, 0)
	}

	var tmp = make(map[string]interface{}, 0)

	u, _ := uuid.NewV4()
	errUUID := u.String()
	tmp["error_uuid"] = errUUID
	tmp["message"] = Wrap(e).Error()
	tmp["context"] = context
	buf, e := json.MarshalIndent(tmp, prefix, indent)
	if e != nil {
		return "", nil, Wrap(e)
	}
	return errUUID, buf, nil
}
