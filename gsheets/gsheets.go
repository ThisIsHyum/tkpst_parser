package gsheets

import (
	"context"
	"slices"
	"strings"
	"time"
	"tkpst_parser/pkg/googleapi"

	"github.com/ThisIsHyum/osago/types"
	"google.golang.org/api/sheets/v4"
)

type GsheetsParser struct {
	srv *sheets.Service
}

func New(ctx context.Context, credentialsPath string) (GsheetsParser, error) {
	srv, err := googleapi.NewSheetsService(ctx, credentialsPath)
	if err != nil {
		return GsheetsParser{}, err
	}
	return GsheetsParser{srv: srv}, nil
}

func (gp GsheetsParser) GetValues(id string) ([]*sheets.RowData, error) {
	resp, err := gp.srv.Spreadsheets.Get(id).Ranges("A1:N2000").IncludeGridData(true).Do()
	if err != nil {
		return nil, err
	}
	return resp.Sheets[0].Data[0].RowData, nil
}

func (gp GsheetsParser) GetGroupNames(values []*sheets.RowData) []string {
	groupNames := []string{}
	for _, value := range values {
		if len(value.Values) == 0 {
			continue
		}
		if s := value.Values[0].FormattedValue; strings.Contains(s, "Группа") {
			groupNames = append(groupNames, strings.ToUpper(strings.TrimPrefix(s, "Группа - ")))
		}
	}
	return groupNames
}

func (gp GsheetsParser) GetLessons(values []*sheets.RowData, groups map[string]uint) ([]types.Lesson, error) {
	var lessons = []types.Lesson{}
	var currentGroup string
	var currentGroupID, currentOrder uint
	for i, value := range values {
		if len(value.Values) == 0 {
			continue
		}

		if s := value.Values[0].FormattedValue; strings.Contains(s, "Группа") {
			currentGroup = strings.ToUpper(strings.TrimPrefix(s, "Группа - "))
			currentGroupID = groups[currentGroup]
			currentOrder = 0
		}

		if len(value.Values) > 2 {
			if strings.Contains(value.Values[2].FormattedValue, "Классный час") {
				continue
			}
		}

		if slices.Contains([]string{"1", "2", "3", "4", "5", "6"}, value.Values[0].FormattedValue) {
			currentOrder++
			for j := range 6 {
				date := getDate(values, j+1)
				lessons = append(lessons, getLesson(
					values[i+2], value.Values, j+1, currentOrder,
					date, currentGroupID,
				),
				)
			}
		}
	}
	return lessons, nil

}

func getLesson(values *sheets.RowData, value []*sheets.CellData, l int, order uint, date time.Time, groupID uint) types.Lesson {
	if len(value) <= l*2 {
		return types.Lesson{}
	}
	cabinet := value[(l*2)+1].FormattedValue
	if len(values.Values) > (l*2)+1 {
		if values.Values[(l*2)+1].FormattedValue != "" {
			cabinet += "/" + values.Values[(l*2)+1].FormattedValue[1:]
		}
	}
	title := value[l*2].FormattedValue
	teacher := strings.Replace(values.Values[l*2].FormattedValue, "\n", "/", 1)
	return types.NewLesson(title, cabinet, teacher, date, groupID, order)
}

func getDate(values []*sheets.RowData, num int) time.Time {
	var strTime string = values[1].Values[2*num].FormattedValue
	if strTime == "" {
		strTime = values[11].Values[2*num].FormattedValue
	}
	strTime = strings.Split(strTime, ", ")[1]
	t, _ := time.Parse("02.01.2006", strTime)
	return t
}
