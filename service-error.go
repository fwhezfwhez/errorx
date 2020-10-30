package errorx

type ServiceError struct {
	Errcode int
	Errmsg  string
}

func (se ServiceError) Error() string {
	return se.Errmsg
}

func (se ServiceError) Equal(dest ServiceError) bool {
	return se.Errmsg == dest.Errmsg && se.Errcode == dest.Errcode
}

func NewServiceError(errmsg string, errcode int) ServiceError {
	return ServiceError{
		Errcode: errcode,
		Errmsg:  errmsg,
	}
}
func IsServiceErr(src error, dest ...error) (ServiceError, bool) {
	for i, _ := range dest {
		if se, ok := isServiceErr(src, dest[i]); ok {
			return se, ok
		}
	}
	return ServiceError{}, false
}
func isServiceErr(src error, dest error) (ServiceError, bool) {
	if src == nil {
		return ServiceError{}, false
	}
	if dest == nil {
		return ServiceError{}, false
	}

	destS, ok := dest.(ServiceError)
	if !ok {
		return ServiceError{}, false
	}

	switch v := src.(type) {
	case Error:
		if v.E.Error() == dest.Error() {
			return destS, true
		}
		return ServiceError{}, false
	case ServiceError:
		if v.Equal(destS) {
			return destS, true
		}
		return v, false
	case error:
		return ServiceError{}, false
	}
	return ServiceError{}, false
}
