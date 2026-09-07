package pkg

func Success(statusCode, msg, data, totalRecords, count, totalPages any) any {
	return map[string]any{
		"data":         data,
		"message":      msg,
		"statusCode":   statusCode,
		"count":        count,
		"totalPages":   totalPages,
		"totalRecords": totalRecords,
	}
}

func Failure(statusCode, msg any) any {
	return map[string]any{
		"message":    msg,
		"statusCode": statusCode,
	}
}
