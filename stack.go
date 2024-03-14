package errorx

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

func PrintStackFormat(flag int, file string, line int, cause string) string {
	var formatGroup = make([]string, 0, 3)
	var formatArgs = make([]interface{}, 0, 3)
	if flag&LdateTime > 0 {
		formatGroup = append(formatGroup, "%s")
		formatArgs = append(formatArgs, time.Now().Format("2006-01-02 15:04:05"))
	}
	if flag&Llongfile > 0 {
		formatGroup = append(formatGroup, "%s")
		trace := fmt.Sprintf("%s:%d", file, line)
		formatArgs = append(formatArgs, trace)
	}
	if flag&LcauseBy > 0 {
		formatGroup = append(formatGroup, "%s")
		formatArgs = append(formatArgs, cause)
	}
	return fmt.Sprintf(strings.Join(formatGroup, " | "), formatArgs...)
}

func stackV2WithMsg(msg string) string {
	_, f, l, _ := runtime.Caller(2)
	rs := fmt.Sprintf("%s %s %s", time.Now().Format("2006/1/2 15:04:05.000"), fmt.Sprintf("%s:%d", f, l), msg)
	return rs
}
func stackV2() string {
	_, f, l, _ := runtime.Caller(2)
	rs := fmt.Sprintf("%s %s", time.Now().Format("2006/1/2 15:04:05.000"), fmt.Sprintf("%s:%d", f, l))
	return rs
}
func stackDepth3() string {
	_, f, l, _ := runtime.Caller(3)
	rs := fmt.Sprintf("%s %s", time.Now().Format("2006/1/2 15:04:05.000"), fmt.Sprintf("%s:%d", f, l))
	return rs
}

type stackline struct {
	trace string
	desc  string
}
