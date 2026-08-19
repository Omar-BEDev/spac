package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"spac/cli"
	"spac/history"
	"spac/network"
	"spac/ui"
)

// This file was generated with the assistance of an artificial intelligence
// coding agent (opencode). UI decisions are commented inline to make the
// reasoning behind them explicit.

const (
	version = "v0.0.1"
	banner  = `
███████╗ ██████╗  █████╗  ██████╗
██╔════╝ ██╔══██╗ ██╔══██╗ ██╔════╝
███████╗ ██████╔╝ ███████║ ██║
╚════██║ ██╔══██╗ ██╔══██║ ██║
███████║ ██║  ██║ ██║  ██║ ╚██████╗
╚══════╝ ╚═╝  ╚═╝ ╚═╝  ╚═╝  ╚═════╝
`
)

func main() {
	// Decision: decorations (blue banner, blue prompt, results, spinner)
	// are enabled only when stdout is a real terminal (see package ui), so
	// piped runs stay plain and portable.
	fmt.Print(ui.Blue(banner))
	fmt.Println("spac ", version, " - interactive HTTP request console")
	fmt.Println(`Type new req "<api link>" [-method(post,get,put,delete)] and press Enter.`)
	fmt.Println(`Command can be repeated with different links or methods. Type "exit" to quit.`)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(ui.Blue(">> "))
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		switch {
		case line == "":
			continue
		case strings.EqualFold(line, "exit"), strings.EqualFold(line, "quit"):
			fmt.Println("bye")
			return
		default:
			handleLine(line)
		}
	}
}

// handleLine runs a single (non-empty) user input line.
func handleLine(line string) {
	if !cli.IsReqCommand(line) {
		fmt.Println(ui.Red("unknown command: " + line))
		return
	}

	req, err := cli.ParseReq(line)
	if err != nil {
		fmt.Println(ui.Red("parse error: " + err.Error()))
		return
	}

	// Decision: each listed method performs its own request and appends its
	// own history entry, so the log reflects every actual network call. A
	// small spinner animates while the request is in flight and is cleared
	// before the result line is printed (no-op when not a terminal).
	for _, method := range req.Methods {
		spinner := ui.NewSpinner(os.Stdout)
		spinner.Start()
		status, err := network.Send(method, req.URL)
		spinner.Stop()

		if err != nil {
			fmt.Println(ui.Red(strings.ToUpper(method) + " " + req.URL + " -> request failed: " + err.Error()))
			continue
		}

		fmt.Println(ui.Blue(strings.ToUpper(method) + " " + req.URL + " -> " + status))

		if err := history.LogAction("new req " + strings.ToUpper(method) + " " + req.URL); err != nil {
			fmt.Println(ui.Red("history: " + err.Error()))
		}
	}
}
