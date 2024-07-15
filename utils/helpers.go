package utils

func AddLogFields(caller string, requestId string) map[string]interface{} {
	fields := map[string]interface{}{
		caller: requestId,
	}
	return fields
}
