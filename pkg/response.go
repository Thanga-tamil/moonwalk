package pkg

func Success(statusCode, msg, data, totalRecords, count any) any {
	return map[string]any{
		"data": data,
		"message": msg,
		"statusCode": statusCode,
		"count": count,
		"totalRecords": totalRecords,
	}
}

func Failure(statusCode, msg any) any {
	return map[string]any{
		"message": msg,
		"statusCode": statusCode,
	}
}
