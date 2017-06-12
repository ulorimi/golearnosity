package learnosity

import "time"

func containsStr(s []string, e string) bool {
	for _, a := range s {
		if e == a {
			return true
		}
	}
	return false
}

func hasKey(s map[string]interface{}, key string) bool {
	if _, ok := s[key]; ok {
		return true
	}
	return false
}

func formatTime(t time.Time) string {
	loc, _ := time.LoadLocation("GMT")
	t = t.In(loc)
	result := t.Format("20060102-1504")
	return result
}
