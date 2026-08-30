package pkg

func Success(statusCode int, msg string, data any, totalRecords, count int) any {
	return map[string]any{
		"data": data,
		"message": msg,
		"statusCode": statusCode,
		"count": count,
		"totalRecords": totalRecords,
	}
}

func Failure(statusCode int, msg string) any {
	return map[string]any{
		"message": msg,
		"statusCode": statusCode,
	}
}
