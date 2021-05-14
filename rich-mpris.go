package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"

	presence "github.com/hugolgst/rich-go/client"
)

const appID = "842495835245510696"
const format = "{{ duration(position) }}/{{ duration(mpris:length) }} ({{ status }})%{{ artist }} - {{ title }}"

func main() {
	presence.Login(appID)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	metadataOut := make(chan string)
	cmd, err := startMPRISNotify(format, metadataOut)
	if err != nil {
		log.Fatalln("mpris error:", err)
	}
	defer cmd.Process.Signal(os.Interrupt)
	var metadata string
	for {
		select {
		case metadata = <-metadataOut:
		case <-interrupt:
			return
		}
		split := strings.SplitN(metadata, "%", 2)
		err = presence.SetActivity(presence.Activity{
			Details:    split[1],
			State:      split[0],
		})
		if err != nil {
			log.Fatalln("rich presence error:", err)
		}
	}
}

func startMPRISNotify(format string, out chan<- string) (*exec.Cmd, error) {
	cmd := exec.Command("playerctl", "-a", "-F", "metadata", "-f", format)

	o, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get playerctl stdout: %w", err)
	}

	if err := cmd.Start(); err != nil {
		o.Close()
		return nil, fmt.Errorf("failed to start playerctl: %w", err)
	}

	go func() {
		defer cmd.Process.Kill()
		defer o.Close()

		var scanner = bufio.NewScanner(o)
		for scanner.Scan() {
			out <- scanner.Text()
		}
	}()

	return cmd, nil
}
