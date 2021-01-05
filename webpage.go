package main

import (
	"bufio"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-ping/ping"
	"github.com/wcharczuk/go-chart"
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

func editHandler(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path[len("/edit/"):]
	p, err := loadPage(title)
	if err != nil {
		p = &Page{Title: title}
	}
	fmt.Fprintf(w, "<h1>Editing %s</h1>"+
		"<form action=\"/safe/%s\" method=\"POST\">"+
		"<textarea name=\"body\">%s</textarea><br>"+
		"<input type=\"submit\" value=\"Save\">"+
		"</form>",
		p.Title, p.Title, p.Body)
}

type Sitelist struct {
	Sitelist string
}

func handler(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintf(w, "Hi there I am a webpage %s", r.URL.Path[1:])
	//sitelist := "bing.com"
	//p, err := loadPage("Pinger")
	//if(err != nil) {
	//    p = &Page{Title: "Pinger"}
	//}
	fmt.Printf("method=%s\n", r.Method)
	if r.Method == "POST" {
		r.ParseForm()
		fmt.Printf("parsed")
		//fmt.Fprintf(w, "r.list = %s", r.FormValue("list"))
		os.Remove("list.txt")
		file, _ := os.Create("list.txt")
		file.Write([]byte(r.FormValue("list")))
		file.Close()
		drawChart(pingit(r.FormValue("list")))
		//for _, stats := range statsList {
		//	fmt.Printf(stats.Addr)
		//}
	}
	file, _ := os.Open("list.txt")
	data := make([]byte, 100)
	count, _ := file.Read(data)
	p := &Sitelist{Sitelist: string(data[:count])}
	t, _ := template.ParseFiles("mainpage.html")
	t.Execute(w, p)
	file.Close()
}

func main() {
	http.HandleFunc("/", handler)
	http.HandleFunc("/edit/", editHandler)
	http.Handle("/assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("./assets"))))
	fmt.Printf("listening...\n")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func drawChart(stats []ping.Statistics) {
	graph := chart.BarChart{
		Title: "Test Bar Chart",
		Background: chart.Style{
			Padding: chart.Box{
				Top: 40,
			},
		},
		Height:   512,
		BarWidth: 60,
		Bars: []chart.Value{
			{Value: 5.25, Label: "Blue"},
			{Value: 4.88, Label: "Green"},
			{Value: 4.74, Label: "Gray"},
			{Value: 3.22, Label: "Orange"},
			{Value: 3, Label: "Test"},
			{Value: 2.27, Label: "??"},
			{Value: 1, Label: "!!"},
		},
	}

	f, _ := os.Create("assets/output.png")
	defer f.Close()
	graph.Render(chart.PNG, f)
}
