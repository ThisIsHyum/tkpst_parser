package replaces

import (
	"strconv"
	"strings"
	"time"

	"github.com/ThisIsHyum/osago/models"
	"github.com/mmonterroca/docxgo/v2/domain"
)

var mapMonths = map[string]time.Month{
	"декабря": time.December, "января": time.January, "февраля": time.February,
	"марта": time.March, "апреля": time.April, "мая": time.May,
	"июня": time.June, "июля": time.July, "августа": time.August,
	"сентября": time.September, "октября": time.October, "ноября": time.November,
}

func ParseReplaces(table domain.Table, groups map[string]int64) ([]*models.DtoReplace, error) {
	var date time.Time
	var dateStr string
	{
		var err error
		dateStr, err = getFirstCellContent(table, 0)
		if err != nil {
			return nil, err
		}
		date = convertToTime(dateStr)
	}

	var replaces = make([]*models.DtoReplace, 0, 4)
	for idx := 3; ; idx++ {
		row, err := table.Row(idx)
		if err != nil {
			return nil, err
		}
		if content, err := getCellContent(row, 1); err != nil {
			return nil, err
		} else if strings.Contains(content, "ЗАМЕН УЧЕБНЫХ ЗАНЯТИЙ") {
			break
		}

		replace, err := parseReplace(row, date, groups)
		if err != nil {
			return nil, err
		}
		if replace != nil {
			replaces = append(replaces, replace...)
		}
	}

	return replaces, nil
}

func parseReplace(row domain.TableRow,
	date time.Time, groups map[string]int64) ([]*models.DtoReplace, error) {
	teacher, err := getCellContent(row, 0)
	if err != nil {
		return nil, err
	}
	if teacher == "" {
		return nil, nil
	}

	ordersStr, err := getCellContent(row, 2)
	if err != nil {
		return nil, err
	}
	ordersSlice := strings.Split(ordersStr, ",")

	var orders = make([]int, len(ordersSlice))
	for i, order := range ordersSlice {
		orders[i], _ = strconv.Atoi(order)
	}

	var replaces = make([]*models.DtoReplace, len(orders))

	groupStr, err := getCellContent(row, 3)
	if err != nil {
		return nil, err
	}
	groupID := groups[groupStr]

	for i, order := range orders {
		replaces[i] = &models.DtoReplace{
			Teacher:        &teacher,
			Date:           date.Format(time.DateOnly),
			Order:          int64(order),
			StudentGroupID: groupID,
		}
	}

	if teacher == "свободно" {
		return replaces, nil
	}

	cabinet, err := getCellContent(row, 4)
	if err != nil {
		return nil, err
	}
	for i := range replaces {
		replaces[i].Cabinet = &cabinet
	}

	return replaces, nil
}

func convertToTime(text string) time.Time {
	dayStr := text[0:2]
	day, _ := strconv.Atoi(dayStr)

	text = text[2:]
	strs := strings.Split(text, " ")

	month := mapMonths[strs[0]]
	year, _ := strconv.Atoi(strs[1])

	return time.Date(year, month, day, 0, 0, 0, 0, time.Local)
}

func getFirstCellContent(table domain.Table, rowIndex int) (string, error) {
	row, err := table.Row(rowIndex)
	if err != nil {
		return "", err
	}
	return getCellContent(row, 0)
}

func getCellContent(row domain.TableRow, index int) (string, error) {
	cell, err := row.Cell(index)
	if err != nil {
		return "", err
	}
	paragraphs := cell.Paragraphs()
	if len(paragraphs) == 0 {
		return "", nil
	}
	return paragraphs[0].Text(), nil
}
