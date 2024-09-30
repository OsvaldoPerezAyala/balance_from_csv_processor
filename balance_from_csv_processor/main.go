package main

import (
	"balance_from_csv_processor/requesthandler"
	"balance_from_csv_processor/utils/refresh"
	"github.com/go-co-op/gocron"
	"github.com/labstack/echo/v4"
	"time"
)

func main() {

	go tickers()
	e := echo.New()

	e.POST("/summary/csv", requesthandler.ProcessCSV)
	e.Logger.Fatal(e.Start(":8080"))
}

func tickers() {

	s2 := gocron.NewScheduler(time.UTC)
	_, _ = s2.Every(30).Minute().StartImmediately().Do(refresh.Task)
	s2.StartAsync()
}
