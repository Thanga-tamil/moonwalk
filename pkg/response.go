package pkg

type Success struct {
	StatusCode   int16
	Data         any
	Message      string
	Count        int16
	TotalPages   int16
	TotalRecords int16
}

func Failure(statusCode, msg any) any {
	return map[string]any{
		"message":    msg,
		"statusCode": statusCode,
	}
}
