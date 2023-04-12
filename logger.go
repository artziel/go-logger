package logger

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/goccy/go-json"
)

type Level string

type Rotation string

const (
	NoRotation      Rotation = ""
	DailyRotation   Rotation = "daily"
	WeeklyRotation  Rotation = "weekly"
	MonthlyRotation Rotation = "monthly"
)

const (
	ErrorLevel   Level = "error"
	WarningLevel Level = "warning"
	InfoLevel    Level = "info"
)

type LogEntry struct {
	Level     Level       `json:"lvl"`
	Tag       string      `json:"tag"`
	Message   string      `json:"message"`
	Timestamp time.Time   `json:"time_stamp"`
	Data      interface{} `json:"data"`
}

type Logger struct {
	locker *sync.Mutex

	Tag      string
	Path     string
	fileName string
}

func (l *Logger) Exists() bool {
	if _, err := os.Stat(l.fileName); err == nil {
		return true
	}

	return false
}

func (l *Logger) Size() (int64, error) {
	if !l.Exists() {
		return 0, nil
	}
	fi, err := os.Stat(l.fileName)
	if err != nil {
		return 0, err
	}
	size := fi.Size()
	return size, nil
}

func (l *Logger) write(entry LogEntry) error {

	entry.Timestamp = time.Now()
	entry.Tag = l.Tag

	output, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	l.locker.Lock()
	defer l.locker.Unlock()

	f, err := os.OpenFile(l.fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.WriteString(string(output) + "\n"); err != nil {
		return err
	}

	return nil
}

func (l *Logger) Error(message string, data interface{}) error {
	return l.write(LogEntry{Level: ErrorLevel, Message: message, Data: data})
}

func (l *Logger) Warning(message string, data interface{}) error {
	return l.write(LogEntry{Level: WarningLevel, Message: message, Data: data})
}

func (l *Logger) Info(message string, data interface{}) error {
	return l.write(LogEntry{Level: InfoLevel, Message: message, Data: data})
}

func (l *Logger) Rotate() error {

	if !l.Exists() {
		return nil
	}

	if size, err := l.Size(); err != nil {
		return err
	} else if size == 0 {
		return nil
	}

	now := time.Now()

	l.locker.Lock()
	defer l.locker.Unlock()

	l.fileName = l.Path + l.Tag + "." + now.Format("20060102") + ".log"

	return nil
}

func New(tag string, path string, rotation Rotation) (*Logger, error) {
	if path != "" && path[len(path)-1:] != "/" {
		path = path + "/"
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil, fmt.Errorf("the path \"%s\" for log files do not exists or is not accesible", path)
	}

	now := time.Now()
	logger := Logger{
		Tag:      tag,
		Path:     path,
		fileName: path + tag + "." + now.Format("20060102") + ".log",
		locker:   &sync.Mutex{},
	}

	if rotation != NoRotation {
		s := gocron.NewScheduler(time.UTC)
		s.WaitForSchedule()

		var schedule *gocron.Scheduler

		switch rotation {
		case "daily":
			schedule = s.Every(1).Day()
		case "weekly":
			schedule = s.Every(1).Week()
		case "monthly":
			schedule = s.Every(1).MonthLastDay()
		default:
			return nil, fmt.Errorf("invalid \"%s\" value for rotation", rotation)
		}
		schedule.At("23:59:59").Do(func() { logger.Rotate() })
		s.StartAsync()
	}

	return &logger, nil
}
