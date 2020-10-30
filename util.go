package errorx

func isBasicErr(e error) bool {
	switch e.(type) {
	case Error:
		return false
	case ServiceError:
		return false
	case error:
		return true
	}
	return false
}
