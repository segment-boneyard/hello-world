package tr

import (
	"encoding/json"
	"fmt"
	"github.com/segment-sources/stripe/api"
	"strconv"
	"strings"
	"time"
)

func GetString(obj map[string]interface{}, key string) string {
	if obj[key] == nil {
		return ""
	}

	if val, ok := obj[key].(string); ok {
		return val
	}

	return ""
}

func GetBool(obj map[string]interface{}, key string) bool {
	if obj[key] == nil {
		return false
	}

	if val, ok := obj[key].(bool); ok {
		return val
	}

	return false
}

func GetStringList(obj map[string]interface{}, key string) []string {
	if obj[key] == nil {
		return nil
	}

	if list, ok := obj[key].([]interface{}); ok {
		var result []string
		for _, item := range list {
			if s, ok := item.(string); ok {
				result = append(result, s)
			}
		}
		return result
	}

	return nil
}

func GetMap(obj map[string]interface{}, key string) map[string]interface{} {
	if obj[key] == nil {
		return nil
	}

	if val, ok := obj[key].(map[string]interface{}); ok {
		return val
	}

	return nil
}

func Flatten(input map[string]interface{}, prefix string, output map[string]interface{}) {
	for key, value := range input {
		if innerMap, ok := value.(map[string]interface{}); ok {
			Flatten(innerMap, fmt.Sprintf("%s%s_", prefix, key), output)
		} else {
			output[prefix+key] = value
		}
	}
}

func GetTimestamp(obj map[string]interface{}, key string) string {
	var number int64
	if number = GetNumber(obj, key); number == 0 {
		return ""
	}

	ts := time.Unix(number, 0).UTC().Format(time.RFC3339)
	ts = strings.Replace(ts, "Z", ".000Z", 1)
	return ts
}

func GetNumber(obj map[string]interface{}, key string) int64 {
	if obj[key] == nil {
		return 0
	}

	if val, ok := obj[key].(json.Number); ok {
		if intVal, err := strconv.ParseInt(string(val), 10, 64); err == nil {
			return intVal
		}
	}

	return 0
}

func GetMapList(obj map[string]interface{}, key string) []map[string]interface{} {
	if obj[key] == nil {
		return nil
	}

	if iList, ok := obj[key].([]interface{}); ok {
		mapList := make([]map[string]interface{}, 0, len(iList))
		for _, i := range iList {
			if m, ok := i.(map[string]interface{}); ok {
				mapList = append(mapList, m)
			}
		}
		return mapList
	}

	return nil
}

func ExtractEventPayload(event api.Object, allowedTypes ...string) api.Object {
	var data map[string]interface{}
	if data = GetMap(event, "data"); data == nil {
		return nil
	}
	var obj map[string]interface{}
	if obj = GetMap(data, "object"); obj == nil {
		return nil
	}
	objType := GetString(obj, "object")

	if len(allowedTypes) > 0 {
		for _, t := range allowedTypes {
			if objType == t {
				return obj
			}
		}
		return nil
	}

	return obj
}
