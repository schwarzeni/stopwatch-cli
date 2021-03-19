package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	var (
		recordSign      = make(chan struct{})
		stopSign        = make(chan struct{})
		resumeSign      = make(chan struct{})
		stopFlag        = false
		interuptSign    = make(chan os.Signal, 1)
		dataChan        = make(chan string)
		printerQuitSign = make(chan struct{})
		startTime       = time.Now()
		ctx, cancel     = context.WithCancel(context.Background())
	)

	// pprof 性能分析
	// _ "net/http/pprof"
	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	go genData(ctx, time.Second, dataChan, recordSign, stopSign, resumeSign)
	go func() {
		printer(dataChan)
		printerQuitSign <- struct{}{}
	}()

	// 监听按键
	go func() {
		for {
			consoleReader := bufio.NewReaderSize(os.Stdin, 1)
			if input, _ := consoleReader.ReadByte(); input == 10 {
				// 回车键
				if stopFlag {
					resumeSign <- struct{}{}
					stopFlag = false
				} else {
					recordSign <- struct{}{}
				}
			}
		}
	}()

	// 监听 ctrl+c
	signal.Notify(interuptSign, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		for {
			<-interuptSign
			if stopFlag {
				cancel()
				return
			} else {
				stopSign <- struct{}{}
				stopFlag = true
			}
		}
	}()

	<-printerQuitSign
	fmtTime := "2006/01/02 15:04:05"
	fmt.Printf("\n%s - %s\n", startTime.Format(fmtTime), time.Now().Format(fmtTime))
}

type status struct {
	timeEscape       time.Duration
	timeEscapeBefore time.Duration
}

func (s status) String() string {
	return fmt.Sprintf("%s (%s)", durationFormatter(s.timeEscape), durationFormatter(s.timeEscapeBefore))
}

type data struct {
	curr    status
	records []status
}

func (d data) String() string {
	var res string
	for idx := range d.records {
		res += fmt.Sprintf("%d. %s\n", idx+1, d.records[idx].String())
	}
	res += "====\n"
	res += d.curr.String()
	return res
}

func printer(dataChan <-chan string) {
	// fmt.Print("\033[H")
	for currData := range dataChan {
		clearCmd := exec.Command("clear")
		clearCmd.Stdout = os.Stdout
		if err := clearCmd.Run(); err != nil {
			log.Println(err)
		}
		// fmt.Print("\033[2J")
		// fmt.Print("\033[H")
		fmt.Println()
		fmt.Print(currData)
	}
}

func genData(ctx context.Context, interval time.Duration, dataChan chan<- string, recordSign <-chan struct{}, stopSign <-chan struct{}, resumeSign <-chan struct{}) {
	var (
		startTime     = time.Now()
		prevTime      = time.Now()
		recordFlag    bool
		records       []status
		ticker        = time.NewTicker(interval)
		quiting       = false
		getCurrStatus = func() status {
			return status{timeEscape: time.Since(startTime), timeEscapeBefore: time.Since(prevTime)}
		}
	)
	defer func() {
		ticker.Stop()
	}()

	for {
		select {
		// 循环中不能使用 time.Tick ，会造成泄漏，加重 CPU 的使用率和耗电
		// 参见 https://pkg.go.dev/time#Tick 和 https://jesseduffield.com/adventures-in-profiling-with-go/
		// case <-time.Tick(interval):
		case <-ticker.C:
			recordFlag = false
		case <-recordSign:
			recordFlag = true
		case <-stopSign:
			stopTime := time.Now()
			dataChan <- (data{curr: getCurrStatus(), records: records}).String() + "\n已暂停，ctrl-c退出，回车键继续"
			select {
			case <-resumeSign:
				dur := time.Since(stopTime)
				prevTime = prevTime.Add(dur)
				startTime = startTime.Add(dur)
			case <-ctx.Done():
				recordFlag = true
				quiting = true
			}
		case <-ctx.Done():
			recordFlag = true
			quiting = true
		}
		curr := getCurrStatus()
		if recordFlag {
			records = append(records, curr)
			prevTime = time.Now()
			curr.timeEscapeBefore = 0
		}
		dataChan <- (data{curr: curr, records: records}).String()
		if quiting {
			close(dataChan)
			return
		}
	}
}

func durationFormatter(d time.Duration) (res string) {
	d = d.Round(time.Second)
	// more than one hour
	if d.Hours() > 1.0 {
		h := d / time.Hour
		res += fmt.Sprintf("%02dh", h)
		d -= h * time.Hour
	}
	// more than one minute
	if d.Minutes() > 1.0 {
		if len(res) != 0 {
			res += ":"
		}
		m := d / time.Minute
		res += fmt.Sprintf("%02dm", m)
		d -= m * time.Minute
	} else {
		if len(res) != 0 {
			res += ":00m"
		}
	}
	if len(res) != 0 {
		res += ":"
	}
	res += fmt.Sprintf("%02ds", d/time.Second)
	return res
}
