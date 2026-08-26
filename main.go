package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"time"
	"tkpst_parser/config"
	"tkpst_parser/parser"

	"github.com/ThisIsHyum/osago"
	"github.com/ThisIsHyum/osago/models"
)

func NewCalls(weekdays []time.Weekday, order int64, begins, ends time.Time) []*models.DtoCall {
	var calls = make([]*models.DtoCall, 0, len(weekdays))
	for _, weekday := range weekdays {
		calls = append(calls, &models.DtoCall{
			Begins:  begins.Format("15:04"),
			Ends:    ends.Format("15:04"),
			Weekday: models.TimeWeekday(weekday),
			Order:   order,
		})
	}
	return calls
}

func hm(hour, minute int) time.Time {
	return time.Date(0, 0, 0, hour, minute, 0, 0, time.UTC)
}

type Parser struct {
	p        parser.Parser
	campuses map[string]string
	groups   map[string][]string

	interval time.Duration
}

func (p *Parser) GetCalls() ([]*models.DtoCall, error) {
	saturday := []time.Weekday{time.Saturday}
	monday := []time.Weekday{time.Monday}
	allDays := []time.Weekday{time.Tuesday, time.Wednesday, time.Thursday, time.Friday}

	calls := []*models.DtoCall{}
	calls = append(calls, NewCalls(monday, 1, hm(8, 00), hm(8, 30))...) // class hour
	calls = append(calls, NewCalls(monday, 2, hm(8, 30), hm(10, 00))...)
	calls = append(calls, NewCalls(monday, 3, hm(10, 10), hm(11, 40))...)
	calls = append(calls, NewCalls(monday, 4, hm(12, 20), hm(13, 50))...)
	calls = append(calls, NewCalls(monday, 5, hm(14, 00), hm(14, 30))...) // class hour
	calls = append(calls, NewCalls(monday, 6, hm(14, 35), hm(16, 05))...)
	calls = append(calls, NewCalls(monday, 7, hm(16, 15), hm(17, 45))...)
	calls = append(calls, NewCalls(monday, 8, hm(17, 50), hm(18, 50))...)

	calls = append(calls, NewCalls(allDays, 1, hm(8, 30), hm(10, 00))...)
	calls = append(calls, NewCalls(allDays, 2, hm(10, 10), hm(11, 40))...)
	calls = append(calls, NewCalls(allDays, 3, hm(12, 20), hm(13, 50))...)
	calls = append(calls, NewCalls(allDays, 4, hm(14, 00), hm(15, 30))...)
	calls = append(calls, NewCalls(allDays, 5, hm(15, 40), hm(17, 10))...)
	calls = append(calls, NewCalls(allDays, 6, hm(17, 20), hm(18, 50))...)

	calls = append(calls, NewCalls(saturday, 1, hm(8, 30), hm(9, 30))...)
	calls = append(calls, NewCalls(saturday, 2, hm(9, 40), hm(10, 40))...)
	calls = append(calls, NewCalls(saturday, 3, hm(10, 50), hm(11, 50))...)
	calls = append(calls, NewCalls(saturday, 4, hm(12, 10), hm(13, 10))...)
	calls = append(calls, NewCalls(saturday, 5, hm(13, 20), hm(14, 20))...)
	calls = append(calls, NewCalls(saturday, 6, hm(14, 30), hm(15, 30))...)
	return calls, nil
}

func (p *Parser) GetStudentGroupNames(campusName string) ([]string, error) {
	return p.groups[campusName], nil
}

func (p *Parser) SendLessons(groups map[string]int64, lessonsChan chan<- []*models.DtoLesson) error {
	for i := 1; ; i++ {
		if i > 1 {
			time.Sleep(p.interval)
		}
		slog.Info("sending lessons started", slog.Any("iter", i))
		for campusName, id := range p.campuses {
			slog.Info("fetching values", slog.Any("campus", campusName))
			values, err := p.p.GetValues(id)
			if err != nil {
				slog.Error("failed to get values", slog.Any("campus", campusName), slog.Any("error", err))
				continue
			}

			slog.Info("getting lessons", slog.Any("campus", campusName))
			lessons, err := p.p.GetLessons(values, groups)
			if err != nil {
				slog.Error("failed to get lessons", slog.Any("campus", campusName), slog.Any("error", err))
				continue
			}

			l := slices.CompactFunc(lessons, func(lesson, _ *models.DtoLesson) bool {
				return lesson.StudentGroupID == 522
			})
			fmt.Println(l[0].Title)

			slog.Info("sending lessons to channel", slog.Any("campus", campusName), slog.Any("lessons_count", len(lessons)))
			lessonsChan <- lessons
		}
		slog.Info("sending lessons ended", slog.Any("iter", i))
	}
}

func (p *Parser) SendReplaces(groups map[string]int64, replacesChan chan<- []*models.DtoReplace) error {
	return nil
}

func NewParser(campuses map[string]string,
	maxRetries int, retryDelay time.Duration, interval time.Duration) (*Parser, error) {
	parser := parser.New(maxRetries, retryDelay)
	groups := map[string][]string{}
	for campusName, campusId := range campuses {
		values, err := parser.GetValues(campusId)
		if err != nil {
			return nil, err
		}
		groups[campusName] = parser.GetGroupNames(values)
	}
	return &Parser{
		campuses: campuses,
		p:        parser,
		groups:   groups,
		interval: interval,
	}, nil
}

func main() {
	ctx := context.Background()
	config, err := config.LoadConfig()
	if err != nil {
		slog.Error("unable load config", "error", err)
		os.Exit(1)
	}
	slog.Info("config is loaded")

	client, err := osago.NewParserClient(ctx, config.OsaUrl, config.Token, config.Timeout)
	if err != nil {
		slog.Error("unable to create parser client", "error", err)
		os.Exit(1)
	}
	slog.Info("client is created", slog.Any("osa_url", config.OsaUrl))

	parser, err := NewParser(map[string]string{
		"Самарцева":               "15yQ7MCTqWIIvfb3Qw9ie-4Off7bpR7ovSwbhXw1Iuyg",
		"Рылеева":                 "11y4dLT68xrStKvDSC7LmCNxrGXhkWsRMjOOYHWYDyc0",
		"Луначарского 1 курс":     "1kTvUP7cH-8l1yJf9cgQ8-hxkk9cPo3N67miH8Nu0abA",
		"Луначарского 2 курс":     "1PqnXQrm84iRwrR8obKysz6UCZXqbxWuDkhc9c2VYAKY",
		"Луначарского 3 и 4 курс": "1h1nu_K66V5KfQDy72rCKXueWSUn84mwl4kQI0aer5B4",
	}, config.MaxRetries, config.RetryDelay, config.Interval)
	if err != nil {
		slog.Error("unable to create parser", "error", err)
		os.Exit(1)
	}
	slog.Info("parser is created")

	client.SetParser(parser)

	var errChan = make(chan error)
	defer close(errChan)
	go func() {
		for err := range errChan {
			slog.Error("client error", slog.Any("error", err))
		}
	}()

	err = client.Run(ctx, errChan)

	if err != nil {
		slog.Error("unable to run parser client", "error", err)
		os.Exit(1)
	}
}
