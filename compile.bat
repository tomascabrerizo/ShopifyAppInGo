@echo off

if not exist .\build mkdir .\build

set CGO_ENABLED=1
go build -o .\build\app.exe .\src\
