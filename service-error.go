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

// IsServiceErr is used to handle service error.
//  Case below will be regarded as service error and return se,true
func IsServiceErr(src error, dest ...error) (ServiceError, bool) {
	if len(dest) == 0 {
		se, ok := src.(ServiceError)
		if ok {
			return se, true
		}

		e, ok2 := src.(Error)
		if ok2 && e.isServiceErr == true {
			return ServiceError{
				Errmsg:  e.serviceErrmsg,
				Errcode: e.serviceErrcode,
			}, true
		}

		return ServiceError{}, false
	}
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

	if src.Error() == dest.Error() && !ok {
		return NewServiceError(src.Error(), 0), true
	}
	if src.Error() == dest.Error() && ok {
		return destS, true
	}

	if !ok {
		if isBasicErr(dest) {
			return isServiceErr(src, NewServiceError(dest.Error(), 0))
		}
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
