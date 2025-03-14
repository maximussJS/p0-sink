package object

func GetNumberFromMap(data map[string]interface{}, key string) (float64, bool) {
	if val, exists := data[key]; exists && val != nil {
		if num, ok := val.(float64); ok {
			return num, true
		}
	}
	return 0, false
}

func GetStringMapFromMap(data map[string]interface{}, key string) (map[string]string, bool) {
	if val, exists := data[key]; exists && val != nil {
		if m, ok := val.(map[string]interface{}); ok {
			result := make(map[string]string)
			for k, v := range m {
				if s, ok := v.(string); ok {
					result[k] = s
				} else {
					return nil, false
				}
			}
			return result, true
		}
	}
	return nil, false
}

func GetStringFromMap(data map[string]interface{}, key string) (string, bool) {
	if val, exists := data[key]; exists && val != nil {
		if s, ok := val.(string); ok {
			return s, true
		}
	}
	return "", false
}

func GetBoolFromMap(data map[string]interface{}, key string) (bool, bool) {
	if val, exists := data[key]; exists && val != nil {
		if b, ok := val.(bool); ok {
			return b, true
		}
	}
	return false, false
}
