package hedging

import (
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

// ///////////////////////////////////////////////////////////////////
// Get system locale cross-platform way
// ///////////////////////////////////////////////////////////////////
func getLocale() (string, error) {
	// Check the LANG environment variable, common on UNIX.
	envlang, ok := os.LookupEnv("LANG")
	if ok {
		return strings.Split(envlang, ".")[0], nil
	}

	// Exec powershell Get-Culture on Windows.
	cmd := exec.Command("powershell", "Get-Culture | select -exp Name")
	output, err := cmd.Output()
	if err == nil {
		return strings.Trim(string(output), "\r\n"), nil
	}

	return "", fmt.Errorf("cannot determine locale")
}

// ///////////////////////////////////////////////////////////////////
// Use system locale to print decimals correctly
// ///////////////////////////////////////////////////////////////////
func GetPrinter() (*message.Printer, error) {
	locale, err := getLocale()
	if err != nil {
		return nil, err
	}
	slog.Debug(fmt.Sprintf("system locale is %s", locale))
	p := message.NewPrinter(language.Make(locale))
	return p, nil
}
