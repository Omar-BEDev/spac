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

package cli

const (
	Method = "method"
	Api    = "api"
	Cmd    = "cmd"
)

var commands = map[string]string{
	"new Req": Cmd,
	"post":    Method,
	"get":     Method,
	"put":     Method,
	"delete":  Method,
}

func IsAvailableCommand(name string) bool {
	_, ok := commands[name]
	if ok {
		return true
	}
	return false
}

// SupportedMethod reports whether name is one of the HTTP methods the console
// can execute (post, get, put, delete). The check is case-sensitive the same
// way the command map is; callers lower-case before asking when needed.
func SupportedMethod(name string) bool {
	return allowedMethods[name]
}
