# spac

**spac** is an interactive HTTP request console written in Go. Type requests in a REPL, get statuses, headers and pretty-printed JSON bodies back instantly — no curl flags to remember.

```
███████╗ ██████╗  █████╗  ██████╗
██╔════╝ ██╔══██╗ ██╔══██╗ ██╔════╝
███████╗ ██████╔╝ ███████║ ██║
╚════██║ ██╔══██╗ ██╔══██║ ██║
███████║ ██║  ██║ ██║  ██║ ╚██████╗
╚══════╝ ╚═╝  ╚═╝ ╚═╝  ╚═╝  ╚═════╝
spac  v0.0.1  - interactive HTTP request console
```

## Features

- **Full line editing** — arrow keys move the cursor and cycle through command history (powered by [chzyer/readline](https://github.com/chzyer/readline)), Backspace/Delete/Home/End work as expected, and `Ctrl-C` cancels the current line without leaving the console.
- **All common HTTP methods** — `get`, `post`, `put`, `delete`, `patch`, `head`, `options`.
- **Custom headers & auth** — send `Authorization: Bearer ...` or any header with `-header(...)` / `-H`.
- **Data-driven body templates** — request bodies live in `templates/body.json`, never hardcoded; pick a structure with `-struct(name)`.
- **Test suites** — run a JSON list of requests and assert expected status codes with `run -tests`.
- **Graceful non-TTY fallback** — piped or redirected input uses a plain scanner, so CI and scripts never hang.
- **History** — submitted commands persist across sessions; every executed request is logged with a date.

## How It Works

```
stdin ──► line reader ──► command parser ──► network sender ──► response display
          (readline or     (cli package)      (network pkg)     (pretty JSON,
           scanner)                                             status, headers)
```

1. The REPL reads a line. On a real terminal it uses the readline editor (cursor movement + history); when stdin is a pipe or file it falls back to `bufio.Scanner`.
2. `cli.ParseReq` / `cli.ParseRunTests` parse and validate the command (URL scheme, methods, headers, struct selector).
3. `network.SendWithBody` performs the request, optionally attaching your headers and a JSON body loaded by the `template` package from `templates/body.json`.
4. The console prints the status line, `Content-Type`, and the response body pretty-printed when it is JSON.
5. Successful actions are appended to `history.log` in your user config directory (`$XDG_CONFIG_HOME/spac/` or `%AppData%\spac\`).

## Download & Build

Requirements: **Go 1.26+**

```bash
# 1. Clone the source code
git clone https://github.com/Omar-BEDev/spac.git
cd spac

# 2. Build the binary
go build -o spac .

# 3. (optional) Install it on your PATH
go install .
```

## Usage & Examples

Start the console:

```bash
./spac
```

### Send a GET request

```
spac>> new req "https://api.github.com" -method(get)
GET https://api.github.com -> 200 OK
content-type: application/json; charset=utf-8
response body:
{
  "current_user_url": "https://api.github.com/user",
  ...
}
```

### POST with a body template and auth header

Given `templates/body.json`:

```json
{
  "struct": {
    "user":    { "name": "user write new name here", "email": "user write email here" },
    "product": { "name": "user write new name here", "price": 0 }
  }
}
```

Send the `user` structure with a bearer token:

```
spac>> new req "https://api.example.com/users" -method(post) -struct(user) -header("Authorization: Bearer <token>")
```

Equivalent with the short header flag:

```
spac>> new req "https://api.example.com/users" -method(post) -H "Authorization: Bearer <token>"
```

### PATCH, HEAD, OPTIONS

```
spac>> new req "https://api.example.com/users/1" -method(patch) -struct(user)
spac>> new req "https://api.example.com/health" -method(head)
spac>> new req "https://api.example.com/" -method(options)
```

### Run a test suite

`tests.json`:

```json
{
  "tests": [
    { "method": "get",  "url": "https://api.example.com/health" },
    { "method": "post", "url": "https://api.example.com/users", "body": { "name": "a" } },
    { "method": "get",  "url": "https://api.example.com/users/999", "expected_status": 404 },
    { "method": "post", "url": "https://api.example.com/items", "body": { "n": 1 }, "status": 201 }
  ]
}
```

```
spac>> run -tests "tests.json"
PASS 1: GET https://api.example.com/health -> 200 OK
PASS 2: POST https://api.example.com/users -> 201 Created
PASS 3: GET https://api.example.com/users/999 -> 404 Not Found
PASS 4: POST https://api.example.com/items -> 201 Created
```

A case with no expected status passes on any `2xx`; with one, only that exact status passes. `expected_status` and its `status` alias are interchangeable.

### Piped (non-interactive) mode

```bash
printf 'new req "https://api.github.com" -method(get)\nexit\n' | ./spac
```

## Project Layout

| Path | Purpose |
|---|---|
| `main.go` | REPL loop, command dispatch, response display |
| `input.go` | Line-reader abstraction: readline editor + plain fallback |
| `cli/` | Command-line parsing (`new req`, `run -tests`) |
| `network/` | HTTP sender returning full responses |
| `template/` | Body-template loader with struct selector |
| `suite/` | Tests-file loading and validation |
| `history/` | Dated action log in the user config directory |
| `ui/` | Terminal colors and spinner (auto-disabled when piped) |

## Contributing

Contributions are welcome! Please read [`CONTRIBUTING.md`](CONTRIBUTING.md) and [`AGENT.md`](AGENT.md) before opening a pull request. The essentials:

1. **Fork & branch** from `dev`:
   ```bash
   git clone https://github.com/<your-fork>/spac.git
   cd spac
   git checkout -b feat/my-feature dev
   ```
2. **Code style** — run `gofmt` on all files, keep function names clear and concise, document exported functions, and comment unusual decisions.
3. **Tests** — add table-driven `go test` coverage for every new feature or fix; all tests must pass:
   ```bash
   go test ./...
   ```
4. **Dependencies** — prefer the Go standard library; if a dependency is required, keep `go.mod`/`go.sum` tidy (`go mod tidy`).
5. **Commits** — use the format:
   ```
   type of commit(what is add or fix or refactor): explain what you do
   ```
   Example: `feat(repl): add readline support for cursor navigation and command history`
6. **Pull requests** — target the **`dev` branch** (PRs to other branches are rejected), follow `.github/PULL_REQUEST_TEMPLATE.md`, and if you used AI tooling, explain what it did and why the change introduces no technical debt.

## License

Apache 2.0 — see [LICENSE](LICENSE).
