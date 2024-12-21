# notice how we avoid spaces in $now to avoid quotation hell in go build command
$now = Get-Date -UFormat "%Y-%m-%d_%T"
$sha1 = (git rev-parse HEAD).Trim()

# Preserve the original path to restore it later
$origPath = $env:Path

# Run go build with specified environment
$env:CGO_ENABLED=1
$env:Path+=";C:\winlibs-mingw\mingw64\bin"
go build -v -o hedging.exe -ldflags "-X main.sha1ver=$sha1 -X main.buildTime=$now"

# Restore the environment
$env:Path = $origPath
Remove-Item Env:\CGO_ENABLED