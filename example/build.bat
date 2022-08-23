@echo off
set CGO_ENABLED=1 
set GOOS=windows
set GOARCH=amd64
go build -ldflags "-w -s -H=windowsgui"
pause