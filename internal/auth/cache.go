package auth

import (
	"os"
	"path/filepath"
	"strconv"
	"time"
)

func getCachePath(context string) string {
	return filepath.Join(os.TempDir(), "ruin-"+context+".lock")
}

func CheckAuthCache(context string, graceSeconds int) bool {
	path := getCachePath(context)
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	age := time.Since(info.ModTime())
	return age < time.Duration(graceSeconds)*time.Second
}

func TouchAuthCache(context string) {
	path := getCachePath(context)
	_ = os.WriteFile(path, []byte(strconv.FormatInt(time.Now().Unix(), 10)), 0644)
}
