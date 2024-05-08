@echo off
set CGO_ENABLED=1 
set GOOS=windows
set GOARCH=x86
go build -ldflags "-w -s -H=windowsgui" -tags="bdebug"
pause