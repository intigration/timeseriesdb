package main

import (
	"bytes"
	"context"
	"html/template"
	"log"
	"net/url"

	"github.com/gorilla/schema"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
)

type ConfigHistorianInflux struct {
	HistorianName string
	Server        string
	Token         string
	Org           string
	Bucket        string
}

func (conf ConfigHistorianInflux) Init(ctx context.Context, histmap map[string]Historian) {
	if conf.HistorianName == "" {
		log.Print("Influx Historian missing a name.")
		return
	}
	h, err := NewHistorianInflux(
		conf.HistorianName,
		conf.Server, // server
		conf.Token,  // token
		conf.Org,    // organization
		conf.Bucket, // bucket
	)
	if err != nil {
		log.Printf("Failure to load historian %s: %v", conf.HistorianName, err)
		return
	}
	histmap[conf.HistorianName] = h
	go h.Run(ctx)
}

func NewHistorianInflux(name, server, token, org, bucket string) (*HistorianInflux, error) {
	h := new(HistorianInflux)
	h.Name = name
	h.c = make(chan []HistorianData, 1024)
	h.Client = influxdb2.NewClient(server, token)
	h.WriteAPI = h.Client.WriteAPI(org, bucket)
	h.Org = org
	h.Bucket = bucket
	h.Server = server
	h.Token = token

	return h, nil
}

// this only stores float64s!!!
type HistorianInflux struct {
	Name     string
	Server   string
	Token    string
	Org      string
	Bucket   string
	WriteAPI api.WriteAPI
	c        chan []HistorianData
	Client   influxdb2.Client
}

func (h *HistorianInflux) Close() {
	log.Printf("Closing Influx Historian %s", h.Name)
}

func (h *HistorianInflux) C() chan<- []HistorianData {
	return h.c
}

func (h *HistorianInflux) Run(ctx context.Context) {
	defer h.Close()

	for {
		select {
		case hd := <-h.c:
			// new data came in so grab it and put it in the format we need for processing
			for i := range hd {
				v := map[string]any{"Value": hd[i].Value}
				p := influxdb2.NewPoint(hd[i].Name, nil, v, hd[i].Timestamp)
				h.WriteAPI.WritePoint(p)
			}

		case <-ctx.Done():
			return
		}
	}
}

func (h *ConfigHistorianInflux) RenderConfig() template.HTML {
	encoder := schema.NewEncoder()

	form := make(map[string][]string)
	err := encoder.Encode(h, form)
	if err != nil {
		return ""
	}

	form2 := make(map[string]string)
	for k := range form {
		form2[k] = form[k][0]
	}

	w := new(bytes.Buffer)
	err = templates.ExecuteTemplate(w, "StructForm.html", form2)
	if err != nil {
		log.Printf("problem with template. %v", err)
		return ""
	}
	return template.HTML(w.String())

}

func (h ConfigHistorianInflux) Name() string {
	return h.HistorianName
}
func (h ConfigHistorianInflux) String() string {
	return h.Name()
}
func (h *ConfigHistorianInflux) Update(form url.Values) error {

	decoder := schema.NewDecoder()
	err := decoder.Decode(h, form)
	if err == nil {
		system.Changes = true
	}

	return err
}

var routerInflux = apiConfigEditor[*ConfigHistorianInflux]{
	ConfTypeName: "Influx DB",
	Path:         "/Historians/Influx",
}

func routerSetupInflux() {
	routerInflux.Init(router)
	routerInflux.Confs = system.WorkingConfig.Historians.Influx
}

func init() {
	subrouters = append(subrouters, routerSetupInflux)
}
