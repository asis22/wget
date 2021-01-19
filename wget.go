package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type ProgressManager struct {
	ContentLength, Downloaded int64
	isEOF                     bool
}

func NewProgressManager(length int64) ProgressManager {
	return ProgressManager{
		length, 0,
		false,
	}
}

func (p *ProgressManager) Write(bytes []byte) (int, error) {
	p.Downloaded += int64(len(bytes))
	return len(bytes), nil
}

func (p *ProgressManager) Print() {
	bars := [12]string{
		"[-- -- -- -- -- -- -- -- -- --]",
		"[XX -- -- -- -- -- -- -- -- --]",
		"[XX XX -- -- -- -- -- -- -- --]",
		"[XX XX XX -- -- -- -- -- -- --]",
		"[XX XX XX XX -- -- -- -- -- --]",
		"[XX XX XX XX XX -- -- -- -- --]",
		"[XX XX XX XX XX XX -- -- -- --]",
		"[XX XX XX XX XX XX XX -- -- --]",
		"[XX XX XX XX XX XX XX XX -- --]",
		"[XX XX XX XX XX XX XX XX XX --]",
		"[XX XX XX XX XX XX XX XX XX XX]",
		"[?? ?? ?? ?? ?? ?? ?? ?? ?? ??]",
	}

	if p.isEOF {
		fmt.Printf("%s Final size: %.2fMB\n",
			bars[10], float64(p.Downloaded) / 1024 / 1024)
	} else if p.ContentLength > 0 {
		i := int(float64(p.Downloaded) / float64(p.ContentLength) * 10)
		fmt.Printf("%s %.2fMB of %.2FMB\n",
			bars[i], float64(p.Downloaded) / 1024 / 1024, float64(p.ContentLength) / 1024 / 1024)
	} else {
		fmt.Printf("%s %.2fMB of ??\n",
			bars[11], float64(p.Downloaded) / 1024 / 1024)
	}

}
func (p *ProgressManager) StartReporting() {
	go func() {
		for !p.isEOF {
			p.Print()
			time.Sleep(time.Second)
		}
	}()
}

func (p *ProgressManager) StopReporting() {
	p.isEOF = true
}

func main() {
	fmt.Print("Enter URL:")

	var URL string
	if _, err := fmt.Scanln(&URL); err != nil {
		panic(err)
	}

	resp, err := http.Get(URL)
	if err != nil {
		println("Error:", err.Error())
		os.Exit(-1)
	} else if resp.StatusCode != http.StatusOK{
		println("Error:", resp.Status)
		os.Exit(-2)
	}

	defer resp.Body.Close()

	fileName := URL[strings.LastIndex(URL, "/")+1 : len(URL)]
	file, err := os.Create(fileName)
	if err != nil {
		println("Error:", err.Error())
		os.Exit(-3)
	}

	defer file.Close()

	manager := NewProgressManager(resp.ContentLength)
	tee := io.TeeReader(resp.Body, &manager)

	manager.StartReporting()
	io.Copy(file, tee)
	manager.StopReporting()

	manager.Print()
}

