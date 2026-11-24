package utils

import "strings"

// ShellQuote wraps a value in single quotes for POSIX shells and fish.
// Embedded single quotes are escaped without leaving the quoted context.
func ShellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", `'\''`) + "'"
}
