package main

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/danomagnum/gologix"
	"github.com/gorilla/schema"
)

type ConfigCIPClass3 struct {
	PLCName      string
	Address      string // IP Address
	Path         string // CIP Path (ex: "1,0" for slot 0 on the backplane)
	Enable       bool
	DefaultRate  time.Duration
	EndpointList []EndpointCIPClass3
}

func (config *ConfigCIPClass3) Init(ctx context.Context, h map[string]Historian) {
	go config.Run(ctx, h)
}

func (config *ConfigCIPClass3) Run(ctx context.Context, h map[string]Historian) {
	var err error
	client := gologix.NewClient(config.Address)
	client.Path, err = gologix.ParsePath(config.Path)
	if err != nil {
		log.Printf("problem starting CipClass3 Client %s: %v", config.PLCName, err)
		return
	}

	// split the endpoints up by poll rate
	poll_groups := make(map[time.Duration][]EndpointCIPClass3)
	for i := range config.EndpointList {
		rate := config.EndpointList[i].Rate
		group, ok := poll_groups[rate]
		if !ok {
			poll_groups[rate] = make([]EndpointCIPClass3, 0)
			group = poll_groups[rate]
		}
		group = append(group, config.EndpointList[i])
		poll_groups[rate] = group
	}

	// Connect
	if config.Enable {
		err = client.Connect()
		if err != nil {
			log.Printf("could not connect to %s: %v", config.PLCName, err)
			return
		}
	}

	subctx, cancel_func := context.WithCancel(ctx)
	// Wait for poll groups
	for k := range poll_groups {
		log.Printf("CIPClass3 %s Rate %v has %d endpoints", config.PLCName, k, len(poll_groups[k]))
		go config.PollGroup(subctx, client, k, poll_groups[k], h)
	}
	<-ctx.Done()
	cancel_func()
	err = client.Disconnect()
	if err != nil {
		log.Printf("problem disconnecting from %s: %v", config.PLCName, err)
	}

}

func (config *ConfigCIPClass3) PollGroup(ctx context.Context, client *gologix.Client, rate time.Duration, endpoints []EndpointCIPClass3, h map[string]Historian) {

	tags := make([]string, len(endpoints))
	types := make([]gologix.CIPType, len(endpoints))

	for i := range endpoints {
		tags[i] = endpoints[i].TagName
		types[i] = endpoints[i].TagType
	}

	t := time.NewTicker(rate)
	for {
		select {
		case <-t.C:
			hd := make([]HistorianData, len(endpoints))
			ts := time.Now()
			values, err := client.ReadList(tags, types)
			if err != nil {
				log.Printf("problem reading %s at %v: %v", config.PLCName, rate, err)
				continue
			}
			for i := range values {
				endpoints[i].Value = values[i]
				hd[i] = HistorianData{
					Timestamp: ts,
					Name:      fmt.Sprintf("%s.%s", config.PLCName, endpoints[i].TagName),
					Value:     values[i],
				}
				if endpoints[i].Historian != "" {
					h[endpoints[i].Historian].C() <- hd
				}
			}
		case <-ctx.Done():
			log.Printf("Closing %s Rate %v.", config.PLCName, rate)
			return
		}
	}
}

type EndpointCIPClass3 struct {
	Name      string
	TagName   string
	Rate      time.Duration
	TagType   gologix.CIPType
	Value     any
	Historian string
}

func (e EndpointCIPClass3) TypeAsInt() int {
	return int(e.TagType)
}

// The functions in this next section are to support the web API
// for editing the config

func (e ConfigCIPClass3) Name() string {
	return e.PLCName
}
func (e ConfigCIPClass3) String() string {
	return e.Name()
}
func (e *ConfigCIPClass3) Update(url.Values) error {
	return nil
}
func (e ConfigCIPClass3) Endpoints() []any {
	eps := make([]any, len(e.EndpointList))
	for i, ep := range e.EndpointList {
		eps[i] = ep
	}
	return eps
}
func (e *ConfigCIPClass3) NewEndpoint() {
	e.EndpointList = append(e.EndpointList, EndpointCIPClass3{Name: "New Endpoint"})

}
func (e *ConfigCIPClass3) UpdateEndpoint(form url.Values) error {

	// see what we're editing.
	index_str := form.Get("Index")
	index, err := strconv.Atoi(index_str)
	if err != nil {
		return fmt.Errorf("invalid endpoint %s not an int: %w", index_str, err)
	} else if index < 0 || index >= len(e.EndpointList) {
		return fmt.Errorf("invalid endpoint %d.  must be 0.. %d", index, len(e.EndpointList))
	}

	// validate the rate
	rate_str := form.Get("Rate")
	rate_str = strings.ReplaceAll(rate_str, " ", "") // get rid of spaces for parsing
	rate, err := time.ParseDuration(rate_str)
	if err != nil {
		return fmt.Errorf("invalid rete %s: %w", rate_str, err)
	}

	// validate the type
	ciptype_str := form.Get("Type")
	ciptype, err := strconv.Atoi(ciptype_str)
	if err != nil {
		return fmt.Errorf("invalid type %s. not an int: %w", ciptype_str, err)
	}

	// load data into that item.
	newendpoint := e.EndpointList[index]
	newendpoint.Historian = form.Get("Historian")
	newendpoint.Name = form.Get("Name")
	newendpoint.TagName = form.Get("TagName")
	newendpoint.Rate = rate
	newendpoint.TagType = gologix.CIPType(ciptype)
	e.EndpointList[index] = newendpoint
	system.Changes = true
	return nil

}
func (e *ConfigCIPClass3) RemoveEndpoint(int) {

}

func (e *ConfigCIPClass3) RenderEndpoints() template.HTML {
	w := new(bytes.Buffer)
	err := templates.ExecuteTemplate(w, "Provider_CIPClass3.html", *e)
	if err != nil {
		log.Printf("problem with template. %v", err)
		return ""
	}
	return template.HTML(w.String())
}

func (e *ConfigCIPClass3) RenderConfig() template.HTML {
	encoder := schema.NewEncoder()

	form := make(map[string][]string)
	err := encoder.Encode(e, form)
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

// Here we create the web API for modifying the CIPClass3 Config.  We've got endpoints so we'll use the apiConfigEditorWithEndpoints
// to handle most of the functionality.  The functions to support this are above.
// We'll also need a function to initialize the routing and update the list of CIPClass3 Instances.
// then we have to register that function as a subrouter.
var routerCIPClass3 = apiConfigEditorWithEndpoints[*ConfigCIPClass3]{
	ConfTypeName: "CIP Class 3",
	Path:         "/Providers/CIPClass3",
	//Confs:        system.WorkingConfig.DataProviders.CIPClass3,
}

func routerSetupCIPClass3() {
	routerCIPClass3.Init(router)
	routerCIPClass3.Confs = system.WorkingConfig.DataProviders.CIPClass3
}

func init() {
	subrouters = append(subrouters, routerSetupCIPClass3)
}
