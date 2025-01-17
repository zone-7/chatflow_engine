package utils

import (
	"strconv"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize"
)

// Div 数字转字母
func ExcelDiv(Num int) string {
	var (
		Str  string = ""
		k    int
		temp []int //保存转化后每一位数据的值，然后通过索引的方式匹配A-Z
	)
	//用来匹配的字符A-Z
	Slice := []string{"", "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O",
		"P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z"}

	if Num > 26 { //数据大于26需要进行拆分
		for {
			k = Num % 26 //从个位开始拆分，如果求余为0，说明末尾为26，也就是Z，如果是转化为26进制数，则末尾是可以为0的，这里必须为A-Z中的一个
			if k == 0 {
				temp = append(temp, 26)
				k = 26
			} else {
				temp = append(temp, k)
			}
			Num = (Num - k) / 26 //减去Num最后一位数的值，因为已经记录在temp中
			if Num <= 26 {       //小于等于26直接进行匹配，不需要进行数据拆分
				temp = append(temp, Num)
				break
			}
		}
	} else {
		return Slice[Num]
	}
	for _, value := range temp {
		Str = Slice[value] + Str //因为数据切分后存储顺序是反的，所以Str要放在后面
	}
	return Str
}

//导入Excel到内存
func ExcelImport(filepath string, fromRow int, toRow int, fromCel int, toCel int) ([][]interface{}, error) {
	if fromRow <= 0 {
		fromRow = 1
	}

	if fromCel <= 0 {
		fromCel = 1
	}

	datas := make([][]interface{}, 0)

	file, err := excelize.OpenFile(filepath)
	if err != nil {
		return datas, err
	}

	sheetName := file.GetSheetName(1)
	rows := file.GetRows(sheetName)

	for r, row := range rows {
		if fromRow > 0 && r+1 < fromRow {
			continue
		}
		if toRow > 0 && r+1 > toRow {
			break
		}

		rowdata := make([]interface{}, 0)
		for c, cel := range row {
			if fromCel > 0 && c+1 < fromCel {
				continue
			}
			if toCel > 0 && c+1 > toCel {
				break
			}

			rowdata = append(rowdata, cel)
		}

		datas = append(datas, rowdata)
	}

	return datas, nil

}

//导出数据到Excel文件
func ExcelExport(filepath string, datas [][]interface{}, fromIndex int, fromCel int) (int, error) {

	file, err := excelize.OpenFile(filepath)
	if err != nil {
		file = excelize.NewFile()
	}

	sheetName := file.GetSheetName(1)
	if fromIndex <= 0 {
		fromIndex = 1
	}
	if fromCel <= 0 {
		fromCel = 1
	}
	//内容
	for _, rowData := range datas {

		for colIndex, colData := range rowData {

			sheetPosition := ExcelDiv(colIndex+fromCel) + strconv.Itoa(fromIndex)
			switch colData.(type) {
			case string:
				file.SetCellValue(sheetName, sheetPosition, colData.(string))
				break
			case int:
				file.SetCellValue(sheetName, sheetPosition, colData.(int))
				break
			case float64:
				file.SetCellValue(sheetName, sheetPosition, colData.(float64))
				break
			case time.Time:
				file.SetCellValue(sheetName, sheetPosition, colData.(time.Time))
			default:
				file.SetCellValue(sheetName, sheetPosition, colData)
			}
		}

		fromIndex += 1
	}

	file.Path = filepath

	file.Save()

	return fromIndex, nil
}
