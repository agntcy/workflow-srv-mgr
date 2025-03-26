// Copyright AGNTCY Contributors (https://github.com/agntcy)
// SPDX-License-Identifier: Apache-2.0
package util

import (
	"fmt"
	"io"
	"os"
	"time"
)

var output io.Writer = os.Stdout

type progressBar struct {
	interval       int
	progressTicker *time.Ticker
	max            int
	count          int
	rendered       bool
}

func newProgressBar(max int, interval int) *progressBar {

	p := &progressBar{interval: interval, max: max}
	return p
}

func (p *progressBar) progress() {

	for range p.progressTicker.C {
		p.count += 1
		p.rendered = true
		if p.count%p.max == 0 {
			fmt.Fprintln(output, "*")
			p.count = 0
		} else {
			fmt.Fprint(output, "*")
		}
	}
}

func (p *progressBar) Start() {

	if p.progressTicker == nil {
		p.progressTicker = time.NewTicker(time.Duration(p.interval) * time.Second)
		go p.progress()
	} else {
		p.progressTicker.Reset(time.Duration(p.interval) * time.Second)
	}
}

func (p *progressBar) Stop() {

	if p.progressTicker != nil {
		p.progressTicker.Stop()

		if p.rendered {
			// Delimitate rendered progress bar with a new line.
			fmt.Fprintln(output)
			p.rendered = false
		}
	}
}

const ErrorPrefix = "-- ERROR: "
const WarningPrefix = "-- WARNING: "
const InfoPrefix = "-- INFO: "

var progressBarOutput = newProgressBar(10, 1)

func SetOutputWriter(writer io.Writer) {
	output = writer
}

func GetOutputWriter() io.Writer {
	return output
}

func OutputError(format string, a ...any) {
	OutputMessage(fmt.Sprint(ErrorPrefix, format), a...)
}

func OutputWarning(format string, a ...any) {
	OutputMessage(fmt.Sprint(WarningPrefix, format), a...)
}

func OutputInfo(format string, a ...any) {
	OutputMessage(fmt.Sprint(InfoPrefix, format), a...)
}

func OutputMessage(format string, a ...any) {
	progressBarOutput.Stop()
	fmt.Fprintf(output, fmt.Sprintln(format), a...)
}

func OutputProgressBar() {
	progressBarOutput.Start()
}

func OutputNewLine() {
	OutputMessage("\n")
}
