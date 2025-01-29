package main

import (
	"context"
	"time"
)

type Historian interface {
	C() chan<- []HistorianData
	Run(context.Context)
}

type HistorianData struct {
	Timestamp time.Time
	Name      string
	Value     any
}
