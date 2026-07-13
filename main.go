package main

import (
	"context"
	"log/slog"
	"os"
	"time"
	"tkpst_parser/config"
	"tkpst_parser/parser"

	"github.com/ThisIsHyum/osago"
	"github.com/ThisIsHyum/osago/types"
)

type Parser struct {
	p        parser.Parser
	campuses map[string]string
	groups   map[string][]string

	interval time.Duration
}

func hm(hour, minute int) time.Time {
	return time.Date(0, 0, 0, hour, minute, 0, 0, time.UTC)
}

func (p *Parser) GetCalls() ([]types.Call, error) {
	saturday := []time.Weekday{time.Saturday}
	allDays := []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday}

	calls := []types.Call{}
	calls = append(calls, types.NewCalls(allDays, 1, hm(8, 30), hm(10, 00))...)
	calls = append(calls, types.NewCalls(allDays, 2, hm(10, 10), hm(11, 40))...)
	calls = append(calls, types.NewCalls(allDays, 3, hm(12, 20), hm(13, 50))...)
	calls = append(calls, types.NewCalls(allDays, 4, hm(14, 00), hm(15, 30))...)
	calls = append(calls, types.NewCalls(allDays, 5, hm(15, 40), hm(17, 10))...)
	calls = append(calls, types.NewCalls(allDays, 6, hm(17, 20), hm(18, 50))...)

	calls = append(calls, types.NewCalls(saturday, 1, hm(8, 30), hm(9, 30))...)
	calls = append(calls, types.NewCalls(saturday, 2, hm(9, 40), hm(10, 40))...)
	calls = append(calls, types.NewCalls(saturday, 3, hm(10, 50), hm(11, 50))...)
	calls = append(calls, types.NewCalls(saturday, 4, hm(12, 10), hm(13, 10))...)
	calls = append(calls, types.NewCalls(saturday, 5, hm(13, 20), hm(14, 20))...)
	calls = append(calls, types.NewCalls(saturday, 6, hm(14, 30), hm(15, 30))...)
	return calls, nil
}

func (p *Parser) GetStudentGroupNames(campusName string) ([]string, error) {
	return p.groups[campusName], nil
}

func (p *Parser) SendLessons(groups map[string]uint, lessonsChan chan<- []types.Lesson) error {
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

			slog.Info("sending lessons to channel", slog.Any("campus", campusName), slog.Any("lessons_count", len(lessons)))
			lessonsChan <- lessons
		}
		slog.Info("sending lessons ended", slog.Any("iter", i))
	}
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

	if err := client.Run(ctx); err != nil {
		slog.Error("unable to run parser client", "error", err)
		os.Exit(1)
	}
}
