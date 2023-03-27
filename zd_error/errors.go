package zd_error

var (
	ParamsErrorCode = 499
	Success         = genError(0, "success")
	ServerError     = genError(500, "内部系统错误")
	ParamsError     = genError(ParamsErrorCode, "参数错误")
	SignError       = genError(501, "签名错误")
)

// 重定义错误码，以便增加新的支持
type ErrorCode struct {
	code Code
	err  error
}

func (c ErrorCode) Unwrap() error {
	return c.err
}

func (c ErrorCode) Error() string {
	return c.code.Error()
}

// Code return error code
func (c ErrorCode) Code() int {
	return c.code.Code()
}

// Message return error message
func (c ErrorCode) Message() string {
	return c.code.Message()
}

// Equal for compatible.
func (c ErrorCode) Equal(err error) bool {
	return c.code.Equal(err)
}

func genError(code int, msg string) ErrorCode {
	return ErrorCode{
		code: Error(code, msg),
	}
}

// 添加错误码，不可重复添加同一错误码
func AddError(code int, msg string) ErrorCode {
	return genError(code, msg)
}

func DMError(err error) Codes {
	if err == nil {
		return Success
	}
	switch err.(type) {
	case ErrorCode:
		return err.(ErrorCode)
	case Code:
		return err.(Code)
	}
	if c, ok := err.(Codes); ok {
		return c
	}
	return ServerError
}
