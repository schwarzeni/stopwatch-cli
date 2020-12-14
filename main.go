package main

import (
	"bufio"
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
		recordSign = make(chan struct{})
		quitSign   = make(chan os.Signal, 1)
		dataChan   = make(chan data)
		startTime  = time.Now()
	)

	go keyPressListener(recordSign)
	go printer(dataChan)
	go genData(time.Second, dataChan, recordSign)

	signal.Notify(quitSign, syscall.SIGINT, syscall.SIGTERM)
	<-quitSign
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
		res += d.records[idx].String() + "\n"
	}
	res += "====\n"
	res += d.curr.String()
	return res
}

func printer(dataChan <-chan data) {
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

func genData(interval time.Duration, dataChan chan<- data, recordSign <-chan struct{}) {
	var (
		startTime  = time.Now()
		prevTime   = time.Now()
		recordFlag bool
		records    []status
	)

	for {
		select {
		case <-time.Tick(interval):
			recordFlag = false
		case <-recordSign:
			recordFlag = true
		}
		curr := status{timeEscape: time.Since(startTime), timeEscapeBefore: time.Since(prevTime)}
		if recordFlag {
			records = append(records, curr)
			prevTime = time.Now()
		}
		dataChan <- data{curr: curr, records: records}
	}
}

func keyPressListener(recordSign chan<- struct{}) {
	for {
		consoleReader := bufio.NewReaderSize(os.Stdin, 1)
		if input, _ := consoleReader.ReadByte(); input == 10 {
			recordSign <- struct{}{}
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