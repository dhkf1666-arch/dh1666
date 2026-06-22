package handlers

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"strings"
	"unicode"

	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
)

type siteStatsColumnMapping struct {
	AccountCol int
	StartCol   int
	EndCol     int
	Source     string
}

type siteStatsColumnOverride struct {
	AccountCol int // 1-based Excel 列号，0 表示未指定
	StartCol   int
	EndCol     int
}

func (m siteStatsColumnMapping) toResponse() gin.H {
	return gin.H{
		"source":      m.Source,
		"account_col": m.AccountCol + 1,
		"start_col":   m.StartCol + 1,
		"end_col":     m.EndCol + 1,
	}
}

func defaultSiteStatsColumns(mode string) siteStatsColumnMapping {
	switch mode {
	case "NN":
		return siteStatsColumnMapping{
			AccountCol: 33,
			StartCol:   29,
			EndCol:     30,
			Source:     "nn-default",
		}
	default:
		return siteStatsColumnMapping{
			AccountCol: 22, // W列(23) 操作人
			StartCol:   18, // S列(19) 开始时间
			EndCol:     19, // T列(20) 完成时间
			Source:     "bx-default",
		}
	}
}

func normalizeSiteStatsHeader(cell string) string {
	cell = strings.TrimSpace(cell)
	cell = strings.Trim(cell, "\ufeff")
	cell = strings.ReplaceAll(cell, " ", "")
	cell = strings.ReplaceAll(cell, "\u00a0", "")
	return strings.ToLower(cell)
}

func findHeaderColumn(header []string, keywords []string) int {
	for _, kw := range keywords {
		nkw := normalizeSiteStatsHeader(kw)
		for i, raw := range header {
			cell := normalizeSiteStatsHeader(raw)
			if cell == "" {
				continue
			}
			if cell == nkw || strings.Contains(cell, nkw) {
				return i
			}
		}
	}
	return -1
}

func resolveSiteStatsColumns(header []string, mode string) siteStatsColumnMapping {
	mapping := defaultSiteStatsColumns(mode)

	accountKeywords := []string{
		"操作人", "后台账号", "操作账号", "操作员账号", "操作员", "会员账号", "账号",
	}
	startKeywords := []string{
		"开始时间", "发起时间", "创建时间", "提交时间",
	}
	endKeywords := []string{
		"完成时间", "结束时间", "处理时间", "到账时间",
	}

	if mode == "NN" {
		accountKeywords = append([]string{"后台账号", "操作账号"}, accountKeywords...)
	}

	if len(header) == 0 {
		return mapping
	}

	if accountCol := findHeaderColumn(header, accountKeywords); accountCol >= 0 {
		mapping.AccountCol = accountCol
		mapping.Source = "header-account"
	}
	if startCol := findHeaderColumn(header, startKeywords); startCol >= 0 {
		mapping.StartCol = startCol
		if mapping.Source == "header-account" {
			mapping.Source = "header-all"
		} else {
			mapping.Source = "header-partial"
		}
	}
	if endCol := findHeaderColumn(header, endKeywords); endCol >= 0 {
		mapping.EndCol = endCol
	}

	log.Printf("[Upload] column mapping mode=%s source=%s account=%d start=%d end=%d",
		mode, mapping.Source, mapping.AccountCol, mapping.StartCol, mapping.EndCol)

	if len(header) > mapping.AccountCol {
		log.Printf("[Upload] header account col sample: [%s]", header[mapping.AccountCol])
	}

	return mapping
}

func resolveSiteStatsColumnMapping(header []string, mode string, override *siteStatsColumnOverride) siteStatsColumnMapping {
	if override != nil && override.AccountCol > 0 && override.StartCol > 0 && override.EndCol > 0 {
		mapping := siteStatsColumnMapping{
			AccountCol: override.AccountCol - 1,
			StartCol:   override.StartCol - 1,
			EndCol:     override.EndCol - 1,
			Source:     "manual",
		}
		log.Printf("[Upload] manual column mapping account=%d start=%d end=%d",
			override.AccountCol, override.StartCol, override.EndCol)
		return mapping
	}
	return resolveSiteStatsColumns(header, mode)
}

func parseSiteStatsColumnOverrideFromForm(c *gin.Context) (*siteStatsColumnOverride, error) {
	accountCol, _ := strconv.Atoi(strings.TrimSpace(c.PostForm("account_col")))
	startCol, _ := strconv.Atoi(strings.TrimSpace(c.PostForm("start_col")))
	endCol, _ := strconv.Atoi(strings.TrimSpace(c.PostForm("end_col")))

	if accountCol == 0 && startCol == 0 && endCol == 0 {
		return nil, nil
	}
	if accountCol <= 0 || startCol <= 0 || endCol <= 0 {
		return nil, fmt.Errorf("请完整选择开始时间、完成时间和操作人列")
	}
	return &siteStatsColumnOverride{
		AccountCol: accountCol,
		StartCol:   startCol,
		EndCol:     endCol,
	}, nil
}

func readSiteStatsRowsFromFileHeader(file *multipart.FileHeader) ([][]string, error) {
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()
	return readSiteStatsRows(src, file.Filename)
}

func readSiteStatsRows(src io.Reader, filename string) ([][]string, error) {
	isCSV := strings.HasSuffix(strings.ToLower(filename), ".csv")
	if isCSV {
		reader := csv.NewReader(src)
		reader.Comma = ','
		reader.FieldsPerRecord = -1
		reader.TrimLeadingSpace = true
		reader.LazyQuotes = true
		rows, err := reader.ReadAll()
		if err != nil {
			return nil, fmt.Errorf("CSV文件解析失败: %w", err)
		}
		return rows, nil
	}

	f, err := excelize.OpenReader(src)
	if err != nil {
		return nil, fmt.Errorf("无法解析 Excel 文件: %w", err)
	}
	defer f.Close()

	sheetName := f.GetSheetName(0)
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func columnNumberToExcelLetter(colNum int) string {
	if colNum <= 0 {
		return ""
	}
	var result strings.Builder
	for colNum > 0 {
		colNum--
		result.WriteByte(byte('A' + colNum%26))
		colNum /= 26
	}
	runes := []rune(result.String())
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func buildSiteStatsColumnOptions(header []string) []gin.H {
	options := make([]gin.H, 0, len(header))
	for i, raw := range header {
		colNum := i + 1
		letter := columnNumberToExcelLetter(colNum)
		headerText := strings.TrimSpace(raw)
		label := fmt.Sprintf("%s列(%d)", letter, colNum)
		if headerText != "" {
			if len(headerText) > 24 {
				headerText = headerText[:24] + "..."
			}
			label = fmt.Sprintf("%s - %s", label, headerText)
		} else {
			label = fmt.Sprintf("%s - (空)", label)
		}
		options = append(options, gin.H{
			"index":  colNum,
			"letter": letter,
			"header": strings.TrimSpace(raw),
			"label":  label,
		})
	}
	return options
}

func padSiteStatsRows(rows [][]string) [][]string {
	maxCols := 0
	for _, row := range rows {
		if len(row) > maxCols {
			maxCols = len(row)
		}
	}
	if maxCols == 0 {
		return rows
	}

	padded := make([][]string, len(rows))
	for i, row := range rows {
		if len(row) >= maxCols {
			padded[i] = row
			continue
		}
		next := make([]string, maxCols)
		copy(next, row)
		padded[i] = next
	}
	return padded
}

func normalizeAccountNameForMatch(name string) string {
	var b strings.Builder
	b.Grow(len(name))

	for _, r := range strings.TrimSpace(name) {
		switch r {
		case '\u00a0', '\u3000', '\ufeff':
			continue
		default:
			if unicode.IsSpace(r) {
				continue
			}
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}

func isLikelyHeaderAccountValue(value string) bool {
	v := normalizeSiteStatsHeader(value)
	switch v {
	case "操作人", "后台账号", "操作账号", "操作员账号", "操作员", "会员账号", "账号",
		"开始时间", "完成时间", "结束时间":
		return true
	default:
		return false
	}
}
