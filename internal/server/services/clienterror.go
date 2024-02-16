package services

type ClientError struct {
	Message string
	Code    int
}

func (c *ClientError) Error() string {
	return c.Message
}

func NewClientError(message string, code int) *ClientError {
	return &ClientError{
		Message: message,
		Code:    code,
	}
}
