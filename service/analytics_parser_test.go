package service_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/sksmagr23/url-shortener-gofr/service"
)

func TestParseUserAgent(t *testing.T) {
	tests := []struct {
		name       string
		ua         string
		expected   service.UserAgentInfo
	}{
		{
			name: "Chrome on Windows",
			ua:   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
			expected: service.UserAgentInfo{
				Browser:    "Chrome",
				OS:         "Windows",
				DeviceType: "Desktop",
			},
		},
		{
			name: "Safari on iPhone",
			ua:   "Mozilla/5.0 (iPhone; CPU iPhone OS 17_0 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/17.0 Mobile/15E148 Safari/604.1",
			expected: service.UserAgentInfo{
				Browser:    "Safari",
				OS:         "iOS",
				DeviceType: "Mobile",
			},
		},
		{
			name: "Firefox on Linux",
			ua:   "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/119.0",
			expected: service.UserAgentInfo{
				Browser:    "Firefox",
				OS:         "Linux",
				DeviceType: "Desktop",
			},
		},
		{
			name: "Edge on Android Tablet",
			ua:   "Mozilla/5.0 (Linux; Android 13; SM-X906B) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/119.0.0.0 Safari/537.36 EdgA/119.0.0.0",
			expected: service.UserAgentInfo{
				Browser:    "Edge",
				OS:         "Android",
				DeviceType: "Tablet",
			},
		},
		{
			name: "Empty UA",
			ua:   "",
			expected: service.UserAgentInfo{
				Browser:    "Unknown",
				OS:         "Unknown",
				DeviceType: "Desktop",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := service.ParseUserAgent(tt.ua)
			assert.Equal(t, tt.expected.Browser, result.Browser)
			assert.Equal(t, tt.expected.OS, result.OS)
			assert.Equal(t, tt.expected.DeviceType, result.DeviceType)
		})
	}
}

func TestExtractIP(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		remote   string
		expected string
	}{
		{
			name:     "X-Forwarded-For",
			headers:  map[string]string{"X-Forwarded-For": "203.0.113.195, 70.41.3.18"},
			remote:   "127.0.0.1:8000",
			expected: "203.0.113.195",
		},
		{
			name:     "X-Real-IP",
			headers:  map[string]string{"X-Real-IP": "203.0.113.196"},
			remote:   "127.0.0.1:8000",
			expected: "203.0.113.196",
		},
		{
			name:     "RemoteAddr Only",
			headers:  map[string]string{},
			remote:   "192.0.2.1:23521",
			expected: "192.0.2.1",
		},
		{
			name:     "RemoteAddr without port",
			headers:  map[string]string{},
			remote:   "192.0.2.1",
			expected: "192.0.2.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest("GET", "http://localhost:8000/test", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			req.RemoteAddr = tt.remote
			ip := service.ExtractIP(req)
			assert.Equal(t, tt.expected, ip)
		})
	}
}

func TestResolveCountry(t *testing.T) {
	assert.Equal(t, "Local", service.ResolveCountry("127.0.0.1"))
	assert.Equal(t, "Local", service.ResolveCountry("::1"))
	assert.Equal(t, "Local", service.ResolveCountry("192.168.1.100"))

	c1 := service.ResolveCountry("8.8.8.8")
	c2 := service.ResolveCountry("8.8.4.4")
	assert.NotEmpty(t, c1)
	assert.NotEmpty(t, c2)
}
