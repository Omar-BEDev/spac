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
