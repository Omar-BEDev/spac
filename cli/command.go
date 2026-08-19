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
