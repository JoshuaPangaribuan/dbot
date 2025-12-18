package logger

import (
	"path/filepath"
	"strconv"
	"strings"
)

func trimmedCaller(file string, line int) string {
	if file == "" {
		return ""
	}

	s := filepath.ToSlash(file)
	parts := strings.Split(s, "/")
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "/" + parts[len(parts)-1] + ":" + strconv.Itoa(line)
	}
	return file + ":" + strconv.Itoa(line)
}
