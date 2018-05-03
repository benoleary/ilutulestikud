# Ilutulestikud #

A backend in Go, a frontend in Angular5-ish.

Getting up and running:
1. install Go
2. `go get github.com/benoleary/ilutulestikud`
3. `cd $GOPATH/src/github.com/benoleary/ilutulestikud` or equivalent for your OS
4. `go install -v`
5. `$GOPATH/bin/ilutulestikud`

After that, you are on your own. Try using `curl` on the endpoints for your `localhost:8081`.

If you are using Go 1.10 or later, the following command runs a full test coverage for all packages in the working directory:
`go test ./... -v -coverprofile=coverage.out ; go tool cover -html=coverage.out`
and that is really convenient.

