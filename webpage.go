package main

import (
	"bufio"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/go-ping/ping"
	"github.com/wcharczuk/go-chart/v2"
)

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
			fmt.Println(stats.Addr)
			fmt.Printf("tx=%d, rx=%d\n", stats.PacketsSent, stats.PacketsRecv)
			fmt.Printf("min=%d, max=%d, avg=%d\n", stats.MinRtt, stats.MaxRtt, stats.AvgRtt)
			waitGroup.Done()
		}()
	}
	waitGroup.Wait()
	return statsList
}

type Page struct {
	Title string
	Body  []byte
}

func (p *Page) save() error {
	filename := p.Title + ".txt"
	return ioutil.WriteFile(filename, p.Body, 0600)
}

func loadPage(title string) (*Page, error) {
	filename := title + ".txt"
	body, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return &Page{Title: title, Body: body}, nil
}

type SiteData struct {
	Sitelist         string
	OutputFileSuffix string
}

func handler(w http.ResponseWriter, r *http.Request) {
	var chartSuffix time.Time
	fmt.Printf("method=%s\n", r.Method)
	if r.Method == "POST" {
		r.ParseForm()
		fmt.Printf("parsed")
		os.Remove("list.txt")
		file, _ := os.Create("list.txt")
		file.Write([]byte(r.FormValue("list")))
		file.Close()
		chartSuffix = drawChart(pingit(r.FormValue("list")))
	} else {
		chartSuffix = time.Unix(0, 0)
	}
	file, _ := os.Open("list.txt")
	data := make([]byte, 100)
	count, _ := file.Read(data)
	p := &SiteData{Sitelist: string(data[:count]), OutputFileSuffix: strconv.FormatInt(chartSuffix.Unix(), 10)}
	t, _ := template.ParseFiles("mainpage.html")
	t.Execute(w, p)
	file.Close()
}

func main() {
	http.HandleFunc("/", handler)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))
	fmt.Printf("listening...\n")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func drawChart(stats []ping.Statistics) time.Time {
	var chartValues []chart.Value
	for _, stat := range stats {
		chartValues = append(chartValues, chart.Value{Value: float64(stat.AvgRtt.Milliseconds()), Label: stat.Addr})
	}
	graph := chart.BarChart{
		Title: "ICMP Ping RTT",
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
