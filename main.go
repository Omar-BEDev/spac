package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"spac/cli"
	"spac/history"
	"spac/network"
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
	fmt.Print(banner)
	fmt.Println("spac ", version, " - interactive HTTP request console")
	fmt.Println(`Type new req "<api link>" [-method(post,get,put,delete)] and press Enter.`)
	fmt.Println(`Command can be repeated with different links or methods. Type "exit" to quit.`)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print(">> ")
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
		fmt.Printf("unknown command: %s\n", line)
		return
	}

	req, err := cli.ParseReq(line)
	if err != nil {
		fmt.Println("parse error:", err)
		return
	}

	// Decision: each listed method performs its own request and appends its
	// own history entry, so the log reflects every actual network call.
	for _, method := range req.Methods {
		status, err := network.Send(method, req.URL)
		if err != nil {
			fmt.Printf("%s %s -> request failed: %v\n", strings.ToUpper(method), req.URL, err)
			continue
		}
		fmt.Printf("%s %s -> %s\n", strings.ToUpper(method), req.URL, status)

		if err := history.LogAction("new req " + strings.ToUpper(method) + " " + req.URL); err != nil {
			fmt.Println("history:", err)
		}
	}
}
