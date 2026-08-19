/*
Copyright 2026 Omar-BEDev

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"spac/cli"
	"spac/history"
	"spac/network"
	"spac/suite"
	"spac/template"
	"spac/ui"
)

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
	fmt.Println(`Type run -tests "<tests.json>" to execute a tests file. Type "exit" to quit.`)

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
	switch {
	case cli.IsRunTestsCommand(line):
		handleRunTests(line)
	case cli.IsReqCommand(line):
		handleNewReq(line)
	default:
		fmt.Println(ui.Red("unknown command: " + line))
	}
}

// handleNewReq runs a "new req" command: one request per listed method, a
// spinner while each request is in flight, and (for POST/PUT only) the JSON
// body template loaded from templates/body.json.
func handleNewReq(line string) {
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
		// Decision: the body is data-driven. For POST and PUT the template is
		// loaded from templates/body.json, displayed to the user so they know
		// what to fill in, and serialized verbatim as the request body. The
		// structure is fixed to the JSON file: no extra fields can be added.
		var body []byte
		if method == "post" || method == "put" {
			body, err = template.LoadForMethod(method, template.DefaultPath)
			if err != nil {
				fmt.Println(ui.Red("template: " + err.Error()))
			}
		}

		spinner := ui.NewSpinner(os.Stdout)
		spinner.Start()
		status, _, err := network.SendWithBody(method, req.URL, body)
		spinner.Stop()

		if err != nil {
			fmt.Println(ui.Red(strings.ToUpper(method) + " " + req.URL + " -> request failed: " + err.Error()))
			continue
		}

		fmt.Println(ui.Blue(strings.ToUpper(method) + " " + req.URL + " -> " + status))
		if len(body) > 0 {
			fmt.Println(ui.Blue("body structure:"))
			fmt.Println(ui.Blue(string(body)))
		}

		if err := history.LogAction("new req " + strings.ToUpper(method) + " " + req.URL); err != nil {
			fmt.Println(ui.Red("history: " + err.Error()))
		}
	}
}

// handleRunTests runs every case in a "run -tests" file, reusing the same
// spinner and history logging as new req. After the request executes, a case
// passes based on the response status:
//
//   - expected_status set: the case passes only for exactly that status;
//   - expected_status omitted: the case passes for any 2xx status.
//
// A transport error (no response at all) is always a FAIL.
func handleRunTests(line string) {
	path, err := cli.ParseRunTests(line)
	if err != nil {
		fmt.Println(ui.Red("run -tests: " + err.Error()))
		return
	}

	cases, err := suite.Load(path)
	if err != nil {
		fmt.Println(ui.Red("run -tests: " + err.Error()))
		return
	}

	for i, tc := range cases {
		spinner := ui.NewSpinner(os.Stdout)
		spinner.Start()
		status, code, err := network.SendWithBody(tc.Method, tc.URL, tc.Body)
		spinner.Stop()

		label := fmt.Sprintf("%d: %s %s", i+1, strings.ToUpper(tc.Method), tc.URL)
		if err != nil {
			fmt.Println(ui.Red("FAIL " + label + " -> " + err.Error()))
			continue
		}

		// Decision: PASS follows the expected status, not just a delivered
		// response. When expected_status is omitted the default is any 2xx,
		// so a 404/500 response prints FAIL instead of a misleading PASS.
		pass := (tc.ExpectedStatus != 0 && code == tc.ExpectedStatus) ||
			(tc.ExpectedStatus == 0 && code >= 200 && code < 300)
		if pass {
			fmt.Println(ui.Blue("PASS " + label + " -> " + status))
		} else if tc.ExpectedStatus != 0 {
			fmt.Println(ui.Red(fmt.Sprintf("FAIL %s -> want status %d got %d (%s)",
				label, tc.ExpectedStatus, code, status)))
		} else {
			fmt.Println(ui.Red(fmt.Sprintf("FAIL %s -> status %d (%s)", label, code, status)))
		}

		// Decision: only genuinely passing cases are recorded in history,
		// matching the existing rule that failed cases leave no trace.
		if pass {
			if err := history.LogAction("run tests " + strings.ToLower(tc.Method) + " " + tc.URL); err != nil {
				fmt.Println(ui.Red("history: " + err.Error()))
			}
		}
	}
}
