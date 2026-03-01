package errors

type Exception struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Err     string `json:"error,omitempty"`
}

func (e *Exception) Error() string {
	return e.Message
}

type UserFriendlyExceptionOption func(*Exception)

func WithCode(code int) UserFriendlyExceptionOption {
	return func(h *Exception) {
		h.Code = code
	}
}

func WithMessage(message string) UserFriendlyExceptionOption {
	return func(h *Exception) {
		h.Message = message
	}
}

func WithError(err error) UserFriendlyExceptionOption {
	return func(h *Exception) {
		h.Err = err.Error()
	}
}

func NotFound(opts ...UserFriendlyExceptionOption) *Exception {
	defaultOpts := []UserFriendlyExceptionOption{
		WithCode(404),
		WithMessage("no entities found with given parameters"),
	}
	return UserFriendlyException(append(defaultOpts, opts...)...)
}

func BadRequest(opts ...UserFriendlyExceptionOption) *Exception {
	defaultOpts := []UserFriendlyExceptionOption{
		WithCode(400),
		WithMessage("bad request"),
	}
	return UserFriendlyException(append(defaultOpts, opts...)...)
}

func Conflict(opts ...UserFriendlyExceptionOption) *Exception {
	defaultOpts := []UserFriendlyExceptionOption{
		WithCode(409),
		WithMessage("conflict"),
	}
	return UserFriendlyException(append(defaultOpts, opts...)...)
}

func Unauthorized(opts ...UserFriendlyExceptionOption) *Exception {
	defaultOpts := []UserFriendlyExceptionOption{
		WithCode(401),
		WithMessage("unauthorized"),
	}
	return UserFriendlyException(append(defaultOpts, opts...)...)
}

func Unexpected(opts ...UserFriendlyExceptionOption) *Exception {
	defaultOpts := []UserFriendlyExceptionOption{
		WithCode(500),
		WithMessage("internal server error"),
	}
	return UserFriendlyException(append(defaultOpts, opts...)...)
}

func UserFriendlyException(opts ...UserFriendlyExceptionOption) *Exception {
	h := &Exception{
		Code:    500,
		Message: "internal server error",
	}
	for _, opt := range opts {
		opt(h)
	}
	return h
}
