package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

func String2sha256(str string) string {

	h := sha256.New()
	h.Write([]byte(str))
	sum := h.Sum(nil)

	s := hex.EncodeToString(sum)
	return s
}

// 将字符串加密成 md5
func String2md5(str string) string {
	data := []byte(str)
	has := md5.Sum(data)
	return fmt.Sprintf("%x", has) //将[]byte转成16进制
}

// RandomString 在数字、大写字母、小写字母范围内生成num位的随机字符串
func RandomString(length int) string {
	// 48 ~ 57 数字
	// 65 ~ 90 A ~ Z
	// 97 ~ 122 a ~ z
	// 一共62个字符，在0~61进行随机，小于10时，在数字范围随机，
	// 小于36在大写范围内随机，其他在小写范围随机
	rand.Seed(time.Now().UnixNano())
	result := make([]string, 0, length)
	for i := 0; i < length; i++ {
		t := rand.Intn(62)
		if t < 10 {
			result = append(result, strconv.Itoa(rand.Intn(10)))
		} else if t < 36 {
			result = append(result, string(rand.Intn(26)+65))
		} else {
			result = append(result, string(rand.Intn(26)+97))
		}
	}
	return strings.Join(result, "")
}

// RandomString 在数字 围内生成num位的随机字符串
func RandomNumber(length int) string {
	// 48 ~ 57 数字
	rand.Seed(time.Now().UnixNano())
	result := make([]string, 0, length)
	for i := 0; i < length; i++ {
		t := rand.Intn(10)
		if t < 10 {
			result = append(result, strconv.Itoa(rand.Intn(10)))
		}
	}
	return strings.Join(result, "")
}

func SubString(str string, begin, length int) string {
	fmt.Println("Substring =", str)
	rs := []rune(str)
	lth := len(rs)
	fmt.Printf("begin=%d, end=%d, lth=%d\n", begin, length, lth)
	if begin < 0 {
		begin = 0
	}
	if begin >= lth {
		begin = lth
	}
	end := begin + length

	if end > lth {
		end = lth
	}
	fmt.Printf("begin=%d, end=%d, lth=%d\n", begin, length, lth)
	return string(rs[begin:end])
}

func GetBetweenStr(str, start, end string) string {
	n := strings.Index(str, start)
	if n == -1 {
		n = 0
	} else {
		n = n + len(start) // 增加了else，不加的会把start带上
	}
	str = string([]byte(str)[n:])
	m := strings.Index(str, end)
	if m == -1 {
		m = len(str)
	}
	str = string([]byte(str)[:m])
	return str
}

// 编辑距离
// Levenshtein levenshtein()
// costIns: Defines the cost of insertion.
// costRep: Defines the cost of replacement.
// costDel: Defines the cost of deletion.
func Levenshtein(str1, str2 string, costIns, costRep, costDel int) int {
	var maxLen = 255
	l1 := len(str1)
	l2 := len(str2)
	if l1 == 0 {
		return l2 * costIns
	}
	if l2 == 0 {
		return l1 * costDel
	}
	if l1 > maxLen || l2 > maxLen {
		return -1
	}

	p1 := make([]int, l2+1)
	p2 := make([]int, l2+1)
	var c0, c1, c2 int
	var i1, i2 int
	for i2 := 0; i2 <= l2; i2++ {
		p1[i2] = i2 * costIns
	}
	for i1 = 0; i1 < l1; i1++ {
		p2[0] = p1[0] + costDel
		for i2 = 0; i2 < l2; i2++ {
			if str1[i1] == str2[i2] {
				c0 = p1[i2]
			} else {
				c0 = p1[i2] + costRep
			}
			c1 = p1[i2+1] + costDel
			if c1 < c0 {
				c0 = c1
			}
			c2 = p2[i2] + costIns
			if c2 < c0 {
				c0 = c2
			}
			p2[i2+1] = c0
		}
		tmp := p1
		p1 = p2
		p2 = tmp
	}
	c0 = p1[l2]

	return c0
}

// 字符串相似度
// SimilarText similar_text()
func SimilarText(first, second string, percent *float64) int {
	var similarText func(string, string, int, int) int
	similarText = func(str1, str2 string, len1, len2 int) int {
		var sum, max int
		pos1, pos2 := 0, 0

		// Find the longest segment of the same section in two strings
		for i := 0; i < len1; i++ {
			for j := 0; j < len2; j++ {
				for l := 0; (i+l < len1) && (j+l < len2) && (str1[i+l] == str2[j+l]); l++ {
					if l+1 > max {
						max = l + 1
						pos1 = i
						pos2 = j
					}
				}
			}
		}

		if sum = max; sum > 0 {
			if pos1 > 0 && pos2 > 0 {
				sum += similarText(str1, str2, pos1, pos2)
			}
			if (pos1+max < len1) && (pos2+max < len2) {
				s1 := []byte(str1)
				s2 := []byte(str2)
				sum += similarText(string(s1[pos1+max:]), string(s2[pos2+max:]), len1-pos1-max, len2-pos2-max)
			}
		}

		return sum
	}

	l1, l2 := len(first), len(second)
	if l1+l2 == 0 {
		return 0
	}

	sim := similarText(first, second, l1, l2)
	if percent != nil {
		*percent = float64(sim*200) / float64(l1+l2)
	}

	return sim
}

func StringsContains(array []string, val string) (index int) {

	for i := 0; i < len(array); i++ {
		if array[i] == val {
			return i
		}
	}
	return -1
}

func StringsRemove(array []string, index int) []string {
	if len(array) == 0 {
		return array
	}

	array = append(array[:index], array[index+1:]...)
	return array
}

// GetStrLength 返回输入的字符串的字数，汉字和中文标点算 1 个字数，英文和其他字符 2 个算 1 个字数，不足 1 个算 1个
func StringLength(str string) int {
	words := strings.Split(str, "")
	return len(words)
}

func InterfaceToString(val interface{}) string {
	var value string

	switch val.(type) {
	case string:
		value = val.(string)
		break
	case int, int8, int32, int64:
		value = fmt.Sprintf("%d", val)
		break
	case uint, uint8, uint32, uint64:
		value = fmt.Sprintf("%d", val)
		break
	case float32, float64:
		value = fmt.Sprintf("%f", val)
		break
	case time.Time:
		value = val.(time.Time).Format("2006-01-02 15:04:05")
		break
	case bool:
		if val.(bool) {
			value = "true"
		} else {
			value = "false"
		}
		break
	case []byte:
		value = string(val.([]byte))
		break

	case []interface{}:
		b, err := json.Marshal(val)
		if err == nil {
			value = string(b)
		}
		break
	default:
		b, err := json.Marshal(val)
		if err == nil {
			value = string(b)
		}
	}

	return value
}

// ToString converts a value to string.
func ToString(value interface{}) string {
	switch value.(type) {
	case string:
		return value.(string)
	case int8:
		return strconv.FormatInt(int64(value.(int8)), 10)
	case int16:
		return strconv.FormatInt(int64(value.(int16)), 10)
	case int32:
		return strconv.FormatInt(int64(value.(int32)), 10)
	case int64:
		return strconv.FormatInt(int64(value.(int64)), 10)
	case uint8:
		return strconv.FormatUint(uint64(value.(uint8)), 10)
	case uint16:
		return strconv.FormatUint(uint64(value.(uint16)), 10)
	case uint32:
		return strconv.FormatUint(uint64(value.(uint32)), 10)
	case uint64:
		return strconv.FormatUint(uint64(value.(uint64)), 10)
	case float32:
		return strconv.FormatFloat(float64(value.(float32)), 'g', -1, 64)
	case float64:
		return strconv.FormatFloat(float64(value.(float64)), 'g', -1, 64)
	case bool:
		return strconv.FormatBool(value.(bool))
	default:
		return fmt.Sprintf("%+v", value)
	}
}
