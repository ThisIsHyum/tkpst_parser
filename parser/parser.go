package parser

import (
	"strings"
	"time"
	"tkpst_parser/pkg/sheets"

	"github.com/ThisIsHyum/osago/types"
)

var lessonNums = map[string]struct{}{
	"1": {}, "2": {}, "3": {}, "4": {}, "5": {}, "6": {},
}

type Parser struct {
	client sheets.Sheets
}

func New(maxRetries int, retryDelay time.Duration) Parser {
	return Parser{
		client: sheets.New(maxRetries, retryDelay),
	}
}

func (p Parser) GetValues(id string) ([][]string, error) {
	return p.client.Values(id)
}

func (p Parser) GetGroupNames(values [][]string) []string {
	groupNames := []string{}
	for _, row := range values {
		if len(row) == 0 {
			continue
		}
		if strings.Contains(row[0], "Группа") {
			groupNames = append(groupNames, strings.ToUpper(strings.TrimPrefix(row[0], "Группа - ")))
		}
	}
	return groupNames
}

func (p Parser) GetLessons(values [][]string, groups map[string]uint) ([]types.Lesson, error) {
	var lessons []types.Lesson
	var currentGroup string
	var currentGroupID, currentOrder uint

	for i, row := range values {
		if len(row) == 0 {
			continue
		}

		if strings.Contains(row[0], "Группа") {
			currentGroup = strings.ToUpper(strings.TrimPrefix(row[0], "Группа - "))
			currentGroupID = groups[currentGroup]
			currentOrder = 0
		}

		if len(row) > 2 && strings.Contains(row[2], "Классный час") {
			continue
		}

		if _, ok := lessonNums[row[0]]; ok {
			currentOrder++
			for j := 0; j < 6; j++ {
				date := getDate(values, j+1)
				if i+2 >= len(values) {
					continue
				}
				lessons = append(lessons, getLesson(
					values[i+2], row, j+1, currentOrder,
					date, currentGroupID,
				))
			}
		}
	}
	return lessons, nil
}

func getLesson(values []string, row []string, l int, order uint, date time.Time, groupID uint) types.Lesson {
	if len(row) <= l*2 {
		return types.Lesson{}
	}

	cabinet := row[(l*2)+1]
	if len(values) > (l*2)+1 && values[(l*2)+1] != "" {
		cabinet += "/" + values[(l*2)+1][1:]
	}

	title := row[l*2]
	teacher := strings.Replace(values[l*2], "\n", "/", 1)

	return types.NewLesson(title, cabinet, teacher, date, groupID, order)
}

func getDate(values [][]string, num int) time.Time {
	if len(values) < 12 {
		return time.Time{}
	}

	strTime := ""
	if len(values[1]) > 2*num {
		strTime = values[1][2*num]
	}
	if strTime == "" && len(values[11]) > 2*num {
		strTime = values[11][2*num]
	}

	parts := strings.Split(strTime, ", ")
	if len(parts) < 2 {
		return time.Time{}
	}

	t, _ := time.Parse("02.01.2006", parts[1])
	return t
}
