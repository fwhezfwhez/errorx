package errorx

func GetStack(e error) []string {
	if e == nil {
		tmp := empty()
		tmp.wrapStackLine("")

		return tmp.Stack()
	}

	switch v := e.(type) {
	case Error:
		return v.Stack()
	}

	tmp := empty()
	tmp.wrapStackLine("")

	return tmp.Stack()
}
