package service

import (
	"net"
	"net/http"
	"strings"
)

// UserAgentInfo represents parsed metadata from a request User-Agent header
type UserAgentInfo struct {
	Browser    string
	OS         string
	DeviceType string
}

// ParseUserAgent extracts Browser, OS, and DeviceType from a raw User-Agent string
func ParseUserAgent(ua string) UserAgentInfo {
	if ua == "" {
		return UserAgentInfo{
			Browser:    "Unknown",
			OS:         "Unknown",
			DeviceType: "Desktop",
		}
	}

	info := UserAgentInfo{
		Browser:    "Other",
		OS:         "Unknown",
		DeviceType: "Desktop",
	}

	// 1. Device Type Detection
	lowerUA := strings.ToLower(ua)
	if strings.Contains(lowerUA, "ipad") {
		info.DeviceType = "Tablet"
	} else if strings.Contains(lowerUA, "android") && !strings.Contains(lowerUA, "mobile") {
		info.DeviceType = "Tablet"
	} else if strings.Contains(lowerUA, "mobile") || strings.Contains(lowerUA, "iphone") || strings.Contains(lowerUA, "ipod") {
		info.DeviceType = "Mobile"
	} else {
		info.DeviceType = "Desktop"
	}

	// 2. OS Detection
	if strings.Contains(ua, "Windows") {
		info.OS = "Windows"
	} else if strings.Contains(ua, "iPhone") || strings.Contains(ua, "iPad") || strings.Contains(ua, "iPod") {
		info.OS = "iOS"
	} else if strings.Contains(ua, "Android") {
		info.OS = "Android"
	} else if strings.Contains(ua, "Macintosh") || strings.Contains(ua, "Intel Mac OS X") {
		info.OS = "macOS"
	} else if strings.Contains(ua, "Linux") {
		info.OS = "Linux"
	}

	// 3. Browser Detection
	if strings.Contains(ua, "Edg/") || strings.Contains(ua, "Edge/") || strings.Contains(ua, "EdgA/") {
		info.Browser = "Edge"
	} else if strings.Contains(ua, "OPR/") || strings.Contains(ua, "Opera") {
		info.Browser = "Opera"
	} else if strings.Contains(ua, "Chrome") {
		// Chrome UA also contains Safari, so Chrome check must be before Safari
		info.Browser = "Chrome"
	} else if strings.Contains(ua, "Firefox") {
		info.Browser = "Firefox"
	} else if strings.Contains(ua, "Safari") && !strings.Contains(ua, "Chrome") {
		info.Browser = "Safari"
	}

	return info
}

// ExtractIP extracts the clean client IP address from request headers, falling back to RemoteAddr
func ExtractIP(r *http.Request) string {
	if r == nil {
		return ""
	}

	// Check X-Forwarded-For for proxy setups
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			cleanIP := strings.TrimSpace(ips[0])
			if cleanIP != "" {
				return cleanIP
			}
		}
	}

	// Check X-Real-IP
	xri := r.Header.Get("X-Real-IP")
	if xri != "" {
		return strings.TrimSpace(xri)
	}

	// Fallback to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}

// ResolveCountry maps an IP address to a realistic country code (ISO-3166-1 alpha-2)
func ResolveCountry(ip string) string {
	if ip == "" || ip == "127.0.0.1" || ip == "::1" || strings.HasPrefix(ip, "192.168.") || strings.HasPrefix(ip, "10.") || strings.HasPrefix(ip, "172.16.") {
		return "Local"
	}

	// Deterministic mapping to simulate Geo-IP behavior
	sum := 0
	for _, char := range ip {
		sum += int(char)
	}

	countries := []string{"US", "IN", "GB", "DE", "FR", "CA", "AU", "JP", "BR", "SG", "NL", "IN", "US", "AE"}
	return countries[sum%len(countries)]
}
