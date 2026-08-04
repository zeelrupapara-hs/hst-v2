package utils

import "strings"

// OperatingSystem names the platform a user agent reports.
//
// The journal shows this beside the channel, so a support desk can tell a trader on a phone from
// one at a desk. Only the name is kept: a full user agent is noise in a table, and the version
// changes under the client without anything happening on the platform.
func OperatingSystem(userAgent string) string {
	ua := strings.ToLower(userAgent)

	switch {
	case ua == "":
		return ""
	// order matters: an iPad reports macintosh too, and android reports linux
	case strings.Contains(ua, "iphone"), strings.Contains(ua, "ipad"), strings.Contains(ua, "ipod"):
		return "iOS"
	case strings.Contains(ua, "android"):
		return "Android"
	case strings.Contains(ua, "windows"):
		return "Windows"
	case strings.Contains(ua, "mac os"), strings.Contains(ua, "macintosh"):
		return "macOS"
	case strings.Contains(ua, "cros"):
		return "ChromeOS"
	case strings.Contains(ua, "linux"), strings.Contains(ua, "x11"):
		return "Linux"
	default:
		return "Unknown"
	}
}
