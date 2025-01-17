package utils

import "strconv"

func Int64ToString(v int64) string {
	s := strconv.FormatInt(v, 10)
	return s
}

func Float64ToString(v float64) string {
	string := strconv.FormatFloat(v, 'f', -1, 64)
	return string
}
func Float32ToString(v float32) string {
	string := strconv.FormatFloat(float64(v), 'f', -1, 32)
	return string
}

// 'b' (-ddddp±ddd，二进制指数)
// 'e' (-d.dddde±dd，十进制指数)
// 'E' (-d.ddddE±dd，十进制指数)
// 'f' (-ddd.dddd，没有指数)
// 'g' ('e':大指数，'f':其它情况)
// 'G' ('E':大指数，'f':其它情况)

func StringToFloat64(v string) (float64, error) {
	float, err := strconv.ParseFloat(v, 64)
	return float, err
}
func StringToFloat32(v string) (float32, error) {
	float, err := strconv.ParseFloat(v, 32)
	return float32(float), err
}

func StringToInt(v string) (int, error) {
	value, err := strconv.Atoi(v)
	return value, err
}

func StringToInt64(v string) (int64, error) {
	i, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return 0, err
	}
	return i, nil
}
