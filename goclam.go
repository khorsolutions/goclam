package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"github.com/getsentry/sentry-go"
)

const (
	ClamscanCmd   = "clamscan"
	SummaryHeader = "----------- SCAN SUMMARY -----------"
)

type ScanResult struct {
	Path      string
	Infected  bool
	Detection string
}

type ParserError struct {
	Err        error
	ClamOutput string
}

func (c ParserError) Error() string {
	return fmt.Sprintf("parsing error: %s", c.Err)
}

func raiseParserError(reason, clamout string) {
	err := ParserError{Err: errors.New(reason), ClamOutput: clamout}
	sentry.CaptureException(err)
	log.Fatal(err)
}

func runCmd(ctx context.Context, cmd string, args ...string) (string, error) {
	var c *exec.Cmd

	if ctx != nil {
		c = exec.CommandContext(ctx, cmd, args...)
	} else {
		c = exec.Command(cmd, args...)
	}

	output, err := c.CombinedOutput()
	if err != nil {
		sentry.CaptureException(err)
		return string(output), err
	}

	if ctx != nil && ctx.Err() != nil {
		sentry.CaptureException(ctx.Err())
		return "", ctx.Err()
	}

	return string(output), nil
}

func parseResultLine(line string) ScanResult {
	if len(line) == 0 {
		raiseParserError("empty result line", line)
	}

	resData := strings.Split(line, ": ")

	infected := true
	detection := ""
	if resData[1] == "OK" {
		infected = false
	} else {
		detection = strings.Replace(resData[1], " FOUND", "", 1)
	}

	return ScanResult{
		Path:      resData[0],
		Infected:  infected,
		Detection: detection,
	}
}

func parseClamOutput(clamout string) []ScanResult {
	scandata := strings.Split(clamout, SummaryHeader)
	if len(scandata) != 2 {
		raiseParserError("got wrong len on summary split", clamout)
	}

	resultLines := strings.Split(strings.TrimSpace(scandata[0]), "\n")
	if len(resultLines) == 0 {
		raiseParserError("empty scan result", clamout)
	}

	results := make([]ScanResult, len(resultLines))
	for i := 0; i < len(results); i++ {
		results[i] = parseResultLine(resultLines[i])
	}

	return results
}
