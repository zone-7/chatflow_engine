package utils

import (
	"encoding/json"
	"os"
	"sort"
	"strings"

	"baliance.com/gooxml/document"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/ledongthuc/pdf"
)

// 读取PDF文件内容
func ReadPdf(filepath string) (string, error) {
	f, r, err := pdf.Open(filepath)
	// remember close file
	defer f.Close()

	if err != nil {
		return "", err
	}

	text_pages := make([]string, 0)
	text_rows := make([]string, 0)
	totalPage := r.NumPage()
	for pageIndex := 1; pageIndex <= totalPage; pageIndex++ {

		p := r.Page(pageIndex)
		if p.V.IsNull() {
			continue
		}

		rows, err := p.GetTextByRow()
		if err != nil {
			continue
		}
		sort.Slice(rows, func(i, j int) bool {
			return rows[i].Position < rows[j].Position

		})
		page_txt := ""
		for _, row := range rows {

			row_txt := ""
			for _, word := range row.Content {
				row_txt += word.S
			}
			text_rows = append(text_rows, row_txt)
			page_txt += row_txt + "\n"
		}

		text_pages = append(text_pages, page_txt)

	}

	return strings.Join(text_pages, "\n"), nil

}

// 读取word 文件内容
func ReadDocx(filepath string) (string, error) {

	doc, err := document.Open(filepath)
	if err != nil {
		return "", err
	}

	text_paragraps := make([]string, 0)

	for _, para := range doc.Paragraphs() {

		//run为每个段落相同格式的文字组成的片段
		text := ""
		for _, run := range para.Runs() {
			p := run.Text()
			text += p

		}
		if len(text) == 0 {
			continue
		}
		text_paragraps = append(text_paragraps, text)
	}

	return strings.Join(text_paragraps, "\n"), nil

}

func ReadTxt(filepath string) (string, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return "", err
	}

	return string(data), err

}

func ReadExcel(filepath string) (string, error) {

	fromRow := 9999999999999
	fromCel := 9999999999999
	toRow := -1
	toCel := -1

	file, err := excelize.OpenFile(filepath)
	if err != nil {
		return "", err
	}

	sheetName := file.GetSheetName(1)
	rows := file.GetRows(sheetName)

	datas := make([][]interface{}, 0)

	for r, row := range rows {
		hasData := false

		for c, cel := range row {
			if len(cel) > 0 {
				hasData = true
				if fromCel > c {
					fromCel = c
				}

				if toCel < c {
					toCel = c
				}

			}
		}

		if hasData && fromRow > r {
			fromRow = r
		}

		if hasData && toRow < r {
			toRow = r
		}
	}

	for _, row := range rows[fromRow:toRow] {

		rowdata := make([]interface{}, 0)
		for _, cel := range row[fromCel:toCel] {

			rowdata = append(rowdata, cel)
		}

		datas = append(datas, rowdata)
	}

	data_byte, err := json.Marshal(datas)

	if err != nil {
		return "", err
	}

	return string(data_byte), nil

}

func ReadFile(filepath string) (string, error) {
	var words string
	var err error
	if strings.Contains(strings.ToLower(filepath), ".pdf") {
		words, err = ReadPdf(filepath)
	}

	if strings.Contains(strings.ToLower(filepath), ".doc") || strings.Contains(strings.ToLower(filepath), ".docx") {
		words, err = ReadDocx(filepath)
	}

	if strings.Contains(strings.ToLower(filepath), ".txt") {
		words, err = ReadTxt(filepath)
	}

	if strings.Contains(strings.ToLower(filepath), ".md") {
		words, err = ReadTxt(filepath)
	}

	if strings.Contains(strings.ToLower(filepath), ".xls") || strings.Contains(strings.ToLower(filepath), ".xlsx") {
		words, err = ReadExcel(filepath)
	}

	return words, err
}
