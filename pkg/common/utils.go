package common

import "strings"

func ConfigStringToMap(in string) map[int]string {
	resMap := make(map[int]string)
	arr := strings.Split(in, ",")
	for i, item := range arr {
		resMap[i] = item
	}
	return resMap
}
