package main

import (
	"context"
	"log"
	"time"
)

type ConfigHistorianLogging struct {
	Name string
}

func (conf ConfigHistorianLogging) Init(ctx context.Context, histmap map[string]Historian) {
	if conf.Name == "" {
		log.Print("Logging Historian missing a name.")
		return
	}
	h, err := NewHistorianLogging()
	h.Name = conf.Name
	if err != nil {
		log.Printf("Failure to load historian %s: %v", conf.Name, err)
		return
	}
	histmap[conf.Name] = h
	go h.Run(ctx)
}

func NewHistorianLogging() (*HistorianLogging, error) {
	h := new(HistorianLogging)
	h.c = make(chan []HistorianData, 1024)
	h.DataCache = map[string][]historianLoggingDataEntry{}
	h.Timeout = time.Minute

	return h, nil
}

type HistorianLogging struct {
	Name      string
	DataCache map[string][]historianLoggingDataEntry
	c         chan []HistorianData
	Timeout   time.Duration
}

func (h *HistorianLogging) Close() {
	log.Printf("Closign Logging Historian %s", h.Name)
}

func (h *HistorianLogging) C() chan<- []HistorianData {
	return h.c
}

func (h *HistorianLogging) Run(ctx context.Context) {
	defer h.Close()

	t := time.NewTicker(h.Timeout)
	for {
		select {
		case hd := <-h.c:
			// new data came in so grab it and put it in the format we need for processing
			for i := range hd {
				dc, ok := h.DataCache[hd[i].Name]
				if !ok {
					dc = make([]historianLoggingDataEntry, 0)
				}
				dc = append(dc, historianLoggingDataEntry{Timestamp: hd[i].Timestamp, Value: hd[i].Value})
				h.DataCache[hd[i].Name] = dc
			}

		case <-t.C:
			// time to write the data out
			log.Printf("Data Cache: %+v", h.DataCache)
			h.DataCache = map[string][]historianLoggingDataEntry{}

		case <-ctx.Done():
			return
		}
	}
}

type historianLoggingDataEntry struct {
	Timestamp time.Time
	Value     any
}
