package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

type ConfigHistorianJSON struct {
	Name     string
	FileName string
}

func (conf ConfigHistorianJSON) Init(ctx context.Context, histmap map[string]Historian) {
	if conf.Name == "" {
		log.Print("JSON Historian missing a name.")
		return
	}
	h, err := NewHistorianJSON(conf.FileName)
	h.Name = conf.Name
	if err != nil {
		log.Printf("Failure to load historian %s: %v", conf.Name, err)
		return
	}
	histmap[conf.Name] = h
	go h.Run(ctx)
}

func NewHistorianJSON(filename string) (*HistorianJSON, error) {
	h := new(HistorianJSON)
	var err error

	h.File, err = os.Create(filename)
	if err != nil {
		return nil, fmt.Errorf("problem opening historian json file %s: %w", filename, err)
	}

	h.JsonEncoder = *json.NewEncoder(h.File)
	h.c = make(chan []HistorianData, 1024)

	return h, nil
}

type HistorianJSON struct {
	Name        string
	File        io.WriteCloser
	JsonEncoder json.Encoder
	c           chan []HistorianData
}

func (h *HistorianJSON) Close() {
	log.Printf("Closing JSON Historian%s", h.Name)
	h.File.Close()
}

func (h *HistorianJSON) C() chan<- []HistorianData {
	return h.c
}

func (h *HistorianJSON) Run(ctx context.Context) {
	defer h.Close()

	for {
		select {
		case hd := <-h.c:
			for i := range hd {
				h.JsonEncoder.Encode(hd[i])
			}
		case <-ctx.Done():
			return

		}
	}
}
