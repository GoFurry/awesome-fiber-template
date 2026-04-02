package common

type Error interface {
	error
	GetErrorCode() int
	GetMsg() string
	GetHTTPStatus() int
}

type appError struct {
	errorCode  int
	httpStatus int
	msg        string
}

func NewError(errorCode, httpStatus int, msg string) *appError {
	return &appError{
		errorCode:  errorCode,
		httpStatus: httpStatus,
		msg:        msg,
	}
}

func NewServiceError(msg string) *appError {
	return NewError(RETURN_FAILED, 500, msg)
}

func (ae appError) Error() string {
	return ae.msg
}

func (ae appError) GetErrorCode() int {
	return ae.errorCode
}

func (ae appError) GetMsg() string {
	return ae.msg
}

func (ae appError) GetHTTPStatus() int {
	return ae.httpStatus
}
