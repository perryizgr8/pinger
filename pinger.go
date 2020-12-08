package main

import (
	"bufio"
	"fmt"
	"github.com/go-ping/ping"
	"os"
	"sync"
	"time"
)

func main() {
	f, _ := os.Open("list.txt")
	scanner := bufio.NewScanner(f)
	var waitGroup sync.WaitGroup
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
			fmt.Println(stats.Addr)
			fmt.Printf("tx=%d, rx=%d\n", stats.PacketsSent, stats.PacketsRecv)
			fmt.Printf("min=%d, max=%d, avg=%d\n", stats.MinRtt, stats.MaxRtt, stats.AvgRtt)
			waitGroup.Done()
		}()
	}
	waitGroup.Wait()
	f.Close()
}
