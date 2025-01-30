package timeseriesdb

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

type ConfigInflux struct {
	HistorianName string
	Server        string
	Token         string
	Org           string
	Bucket        string
}

func (config ConfigInflux) Init(ctx context.Context, histmap map[string]Historian) {
	if config.HistorianName == "" {
		log.Print("Influx Historian missing a name.")
		return
	}
	h, err := NewInflux(
		config.HistorianName,
		config.Server, // server
		config.Token,  // token
		config.Org,    // organization
		config.Bucket, // bucket
	)
	if err != nil {
		log.Printf("Failure to load historian %s: %v", config.HistorianName, err)
		return
	}
	histmap[config.HistorianName] = h
	go h.Run(ctx)
	log.Printf("Historian Connected %s: %v", config.HistorianName, h.Client)
}

func NewInflux(name, server, token, org, bucket string) (*INflux, error) {
	h := new(INflux)
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
type INflux struct {
	Name     string
	Server   string
	Token    string
	Org      string
	Bucket   string
	WriteAPI api.WriteAPI
	c        chan []HistorianData
	Client   influxdb2.Client
}

func (h *INflux) Close() {
	log.Printf("Closing Influx Historian %s", h.Name)
}

func (h *INflux) C() chan<- []HistorianData {
	return h.c
}

func (h *INflux) Run(ctx context.Context) {
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
func (e *ConfigCIPClass3) RenderPOints() template.HTML {
	w := new(bytes.Buffer)
	err := templates.ExecuteTemplate(w, "Provider_Influx.html", *e)
	if err != nil {
		log.Printf("problem with template. %v", err)
		return ""
	}
	return template.HTML(w.String())
}
func (h *ConfigInflux) RenderConfig() template.HTML {
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

func (h ConfigInflux) Name() string {
	return h.HistorianName
}
func (h ConfigInflux) String() string {
	return h.Name()
}
func (h *ConfigInflux) Update(form url.Values) error {

	decoder := schema.NewDecoder()
	err := decoder.Decode(h, form)
	if err == nil {
		system.Changes = true
	}

	return err
}

var crouter = apiConfigEditor[*ConfigInflux]{
	ConfTypeName: "Influx DB",
	Path:         "/Providers/Data",
}

func rInflux() {
	crouter.Init(router)
	crouter.Confs = system.WorkingConfig.DataProviders.InfluxData
}

func init() {
	subrouters = append(subrouters, rInflux)
}
