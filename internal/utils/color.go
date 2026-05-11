package utils

import (
	"crypto/md5" //nolint:gosec // no need strong crypto here
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
)

var hexColorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func DeriveHexColor(seed string) string {
	sum := md5.Sum([]byte(strings.ToLower(strings.TrimSpace(seed)))) //nolint:gosec // no need strong crypto here
	return fmt.Sprintf("#%s", hex.EncodeToString(sum[:3]))
}

func NormalizeHexColor(value string) (string, bool) {
	value = strings.TrimSpace(value)
	if !hexColorPattern.MatchString(value) {
		return "", false
	}
	return strings.ToLower(value), true
}
