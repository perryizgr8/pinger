package main

import (
	"bufio"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-ping/ping"
	"github.com/wcharczuk/go-chart/v2"
)

type siteData struct {
	Sitelist         string
	OutputFileSuffix string
}

type page struct {
	Title string
	Body  []byte
}

func pingit(list string) []ping.Statistics {
	var statsList []ping.Statistics
	var waitGroup sync.WaitGroup
	scanner := bufio.NewScanner(strings.NewReader(list))
	for scanner.Scan() {
		line := scanner.Text()
		waitGroup.Add(1)
		go func() {
			fmt.Printf("pinging %s\n", line)
			pinger, err := ping.NewPinger(line)
			if err != nil {
				panic(err)
			}
			pinger.Count = 3
			pinger.Timeout = time.Second * 10
			err = pinger.Run()
			if err != nil {
				panic(err)
			}
			stats := pinger.Statistics()
			statsList = append(statsList, *stats)
			waitGroup.Done()
		}()
	}
	waitGroup.Wait()
	return statsList
}

func (p *page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &page{Title: title, Body: body}, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	var chartSuffix time.Time
	fmt.Printf("method=%s\n", r.Method)
	if r.Method == "POST" {
		r.ParseForm()
		os.Remove("list.txt")
		file, _ := os.Create("list.txt")
		file.Write([]byte(r.FormValue("list")))
		file.Close()
		chartSuffix = drawChart(pingit(r.FormValue("list")))
		deleteOldCharts(chartSuffix)
	} else {
		chartSuffix = time.Unix(0, 0)
	}
	file, _ := os.Open("list.txt")
	data := make([]byte, 100)
	count, _ := file.Read(data)
	p := &siteData{Sitelist: string(data[:count]), OutputFileSuffix: strconv.FormatInt(chartSuffix.Unix(), 10)}
	t, _ := template.ParseFiles("mainpage.html")
	t.Execute(w, p)
	file.Close()
}

func deleteOldCharts(newChart time.Time) {
	var oldCharts []string
	dir := "assets"
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		oldCharts = append(oldCharts, path)
		return nil
	})
	fmt.Println("list:")
	for _, oldChart := range oldCharts {
		if (oldChart != "assets") && (oldChart != fmt.Sprintf("assets/output%d.png", newChart.Unix())) && (oldChart != "assets/output0.png") {
			fmt.Println("deleting " + oldChart)
			os.Remove(oldChart)
		}
	}
}

func main() {
	http.HandleFunc("/", handler)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))
	fmt.Printf("listening...\n")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func drawChart(stats []ping.Statistics) time.Time {
	var max float64
	var chartValues []chart.Value
	for _, stat := range stats {
		chartValues = append(chartValues, chart.Value{Value: float64(stat.AvgRtt.Milliseconds()), Label: stat.Addr})
		max = math.Max(max, float64(stat.AvgRtt.Milliseconds()))
	}
	graph := chart.BarChart{
		Title: "ICMP Ping RTT (milliseconds)",
		YAxis: chart.YAxis{
			Range: &chart.ContinuousRange{
				Min: 0.0,
				Max: max,
			},
		},
		Background: chart.Style{
			Padding: chart.Box{
				Top: 40,
			},
		},
		Height:   512,
		BarWidth: 60,
		Bars:     chartValues,
	}
	chartSuffix := time.Now()
	f, _ := os.Create(fmt.Sprintf("assets/output%d.png", chartSuffix.Unix()))
	defer f.Close()
	graph.Render(chart.PNG, f)
	return chartSuffix
}
