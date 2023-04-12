package main

import (
	"github.com/artziel/go-logger"
)

func main() {

	l, err := logger.New("logger", "./", logger.DailyRotation)

	if err != nil {
		panic(err)
	}

	l.Error("This is an Error entry", nil)
	l.Info("This is an Info entry", map[string]interface{}{
		"username": "alan@alus.com.mx",
		"fullname": "El Alan",
		"id":       1,
	})
	l.Warning("This is a Warning entry", nil)

}
