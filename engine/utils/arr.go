package utils

func StringsIndex(arr []string, str string) int {
	for index, value := range arr {
		if value == str {
			return index
		}
	}
	return -1
}
