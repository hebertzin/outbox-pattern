package infra

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
		WithMessage("No entities found with given parameters"),
	}

	defaultOpts = append(defaultOpts, opts...)

	return UserFriendlyException(defaultOpts...)
}

func BadRequest(opts ...UserFriendlyExceptionOption) *Exception {
	defaultOpts := []UserFriendlyExceptionOption{
		WithCode(400),
		WithMessage("bad request"),
	}

	defaultOpts = append(defaultOpts, opts...)

	return UserFriendlyException(defaultOpts...)
}

func Confilct(opts ...UserFriendlyExceptionOption) *Exception {
	defaultOpts := []UserFriendlyExceptionOption{
		WithCode(409),
		WithMessage("conflict"),
	}

	defaultOpts = append(defaultOpts, opts...)

	return UserFriendlyException(defaultOpts...)
}

func Unauthorized(opts ...UserFriendlyExceptionOption) *Exception {
	defaultOpts := []UserFriendlyExceptionOption{
		WithCode(401),
		WithMessage("You need to login first"),
	}

	defaultOpts = append(defaultOpts, opts...)

	return UserFriendlyException(defaultOpts...)
}

func Unexpected(opts ...UserFriendlyExceptionOption) *Exception {
	defaultOpts := []UserFriendlyExceptionOption{
		WithCode(500),
		WithMessage("Now you could PANIC! We don't have any idea what's happening here"),
	}

	defaultOpts = append(defaultOpts, opts...)

	return UserFriendlyException(defaultOpts...)
}

func UserFriendlyException(opts ...UserFriendlyExceptionOption) *Exception {
	const (
		defaultMessage = "This is a friendly error, don't panic! Everthing is under control"
		defaultCode    = 500
	)

	h := &Exception{
		Code:    defaultCode,
		Message: defaultMessage,
	}

	for _, opt := range opts {
		opt(h)
	}

	return h
}
