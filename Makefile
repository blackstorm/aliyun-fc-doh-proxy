build:
	GOOS=linux CGO_ENABLED=0 go build main.go
	zip fc-golang.zip main