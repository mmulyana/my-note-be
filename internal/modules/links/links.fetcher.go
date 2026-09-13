package links

import (
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

const (
	// metadata ada di <head>; lebih dari setengah MB cuma body yang bakal dibuang
	maxBodyBytes = 512 << 10
	maxRedirects = 3
	fetchTimeout = 8 * time.Second
	userAgent    = "my-note-linkbot/1.0 (+link preview)"
)

var (
	ErrUnsupportedScheme = errors.New("only http and https links can be previewed")
	ErrBlockedHost       = errors.New("that address is not reachable")
	ErrFetchFailed       = errors.New("could not read that page")
)

// SSRF: yang dicek IP di dialer, bukan hostname biar DNS rebinding nggak lolos, diulang tiap redirect
var client = &http.Client{
	Timeout: fetchTimeout,
	Transport: &http.Transport{
		DialContext:           safeDial,
		TLSHandshakeTimeout:   5 * time.Second,
		ResponseHeaderTimeout: 5 * time.Second,
		DisableKeepAlives:     true,
	},
	CheckRedirect: func(req *http.Request, via []*http.Request) error {
		if len(via) >= maxRedirects {
			return ErrFetchFailed
		}
		return guardScheme(req.URL)
	},
}

func safeDial(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}

	ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, ErrBlockedHost
	}

	dialer := &net.Dialer{Timeout: 5 * time.Second}
	lastErr := error(ErrBlockedHost)
	for _, ip := range ips {
		if !isPublicIP(ip.IP) {
			continue
		}
		// dial ke IP yang udah dicek, biar resolusi kedua nggak bisa nyelundupin IP lain
		conn, err := dialer.DialContext(ctx, network, net.JoinHostPort(ip.IP.String(), port))
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func isPublicIP(ip net.IP) bool {
	if ip == nil ||
		ip.IsUnspecified() ||
		ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsInterfaceLocalMulticast() ||
		ip.IsMulticast() {
		return false
	}

	if v4 := ip.To4(); v4 != nil {
		switch {
		case v4[0] == 100 && v4[1] >= 64 && v4[1] <= 127: // carrier-grade NAT
			return false
		case v4[0] == 192 && v4[1] == 0 && v4[2] == 0: // IETF protocol assignments
			return false
		case v4[0] == 198 && (v4[1] == 18 || v4[1] == 19): // benchmarking
			return false
		}
		return true
	}

	return ip.IsGlobalUnicast()
}

func guardScheme(u *url.URL) error {
	if u.Scheme != "http" && u.Scheme != "https" {
		return ErrUnsupportedScheme
	}
	if u.Hostname() == "" {
		return ErrUnsupportedScheme
	}
	return nil
}

// fetchPreview nggak pernah gagal separo: halaman yang nggak kebaca tetap balik preview dari URL
func fetchPreview(raw string) (*Preview, error) {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil {
		return nil, ErrUnsupportedScheme
	}
	if err := guardScheme(u); err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, ErrUnsupportedScheme
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml;q=0.9")
	req.Header.Set("Accept-Language", "en;q=0.8")

	resp, err := client.Do(req)
	if err != nil {
		if errors.Is(err, ErrBlockedHost) || strings.Contains(err.Error(), ErrBlockedHost.Error()) {
			return nil, ErrBlockedHost
		}
		return nil, ErrFetchFailed
	}
	defer resp.Body.Close()

	// URL final setelah redirect: acuan og:image relatif, sekaligus target link card
	final := resp.Request.URL

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, ErrFetchFailed
	}

	contentType := resp.Header.Get("Content-Type")
	if ct := strings.ToLower(contentType); ct != "" &&
		!strings.Contains(ct, "text/html") && !strings.Contains(ct, "xhtml") {
		// gambar atau PDF nggak punya <head> — cuma URL yang bisa dipakai
		return fallbackPreview(final), nil
	}

	body := io.LimitReader(resp.Body, maxBodyBytes)
	reader, err := charset.NewReader(body, contentType)
	if err != nil {
		reader = body
	}

	doc, err := html.Parse(reader)
	if err != nil {
		return fallbackPreview(final), nil
	}

	return buildPreview(doc, final), nil
}

func fallbackPreview(u *url.URL) *Preview {
	host := strings.TrimPrefix(u.Hostname(), "www.")
	return &Preview{
		URL:      u.String(),
		Title:    host,
		Favicon:  defaultFavicon(u),
		SiteName: host,
	}
}

func defaultFavicon(u *url.URL) string {
	return (&url.URL{Scheme: u.Scheme, Host: u.Host, Path: "/favicon.ico"}).String()
}
