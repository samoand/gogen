package typeutil

import "strconv"

func BoolToString(value interface{}) string {
	firsttryasstr, ok := value.(string)
	if ok {
		return firsttryasstr
	}
	return strconv.FormatBool(value.(bool))
}
