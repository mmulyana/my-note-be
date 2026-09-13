package links

import (
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

// buildPreview jalan sekali: OpenGraph menang, lalu Twitter card, terakhir <title>/<meta description>
func buildPreview(doc *html.Node, base *url.URL) *Preview {
	var (
		ogTitle, ogDesc, ogImage, ogSite string
		twTitle, twDesc, twImage         string
		docTitle, metaDesc, iconHref     string
		iconRank                         int
	)

	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "title":
				if docTitle == "" {
					docTitle = strings.TrimSpace(textOf(n))
				}
			case "meta":
				key := strings.ToLower(attr(n, "property"))
				if key == "" {
					key = strings.ToLower(attr(n, "name"))
				}
				content := strings.TrimSpace(attr(n, "content"))
				if content == "" {
					break
				}
				switch key {
				case "og:title":
					ogTitle = content
				case "og:description":
					ogDesc = content
				case "og:image", "og:image:url", "og:image:secure_url":
					if ogImage == "" {
						ogImage = content
					}
				case "og:site_name":
					ogSite = content
				case "twitter:title":
					twTitle = content
				case "twitter:description":
					twDesc = content
				case "twitter:image", "twitter:image:src":
					if twImage == "" {
						twImage = content
					}
				case "description":
					metaDesc = content
				}
			case "link":
				rel := strings.ToLower(attr(n, "rel"))
				if !strings.Contains(rel, "icon") || strings.Contains(rel, "mask-icon") {
					break
				}
				href := strings.TrimSpace(attr(n, "href"))
				if href == "" {
					break
				}
				// halaman sering punya banyak ikon; ambil yang paling gede, card-nya 13px di retina
				if rank := iconSizeRank(attr(n, "sizes"), rel); rank >= iconRank {
					iconRank = rank
					iconHref = href
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)

	host := strings.TrimPrefix(base.Hostname(), "www.")

	p := &Preview{
		URL:         base.String(),
		Title:       firstNonEmpty(ogTitle, twTitle, docTitle, host),
		Description: firstNonEmpty(ogDesc, twDesc, metaDesc),
		Image:       resolve(base, firstNonEmpty(ogImage, twImage)),
		Favicon:     resolve(base, iconHref),
		SiteName:    firstNonEmpty(ogSite, host),
	}
	if p.Favicon == "" {
		p.Favicon = defaultFavicon(base)
	}
	return p
}

func iconSizeRank(sizes, rel string) int {
	sizes = strings.ToLower(strings.TrimSpace(sizes))
	if sizes == "" {
		if strings.Contains(rel, "apple-touch-icon") {
			return 180
		}
		return 1
	}
	if sizes == "any" {
		return 512
	}
	// "32x32 16x16" angka pertama udah cukup buat ngurutin
	if idx := strings.IndexByte(sizes, 'x'); idx > 0 {
		n := 0
		for _, r := range sizes[:idx] {
			if r < '0' || r > '9' {
				return 1
			}
			n = n*10 + int(r-'0')
		}
		return n
	}
	return 1
}

func attr(n *html.Node, name string) string {
	for _, a := range n.Attr {
		if strings.EqualFold(a.Key, name) {
			return a.Val
		}
	}
	return ""
}

func textOf(n *html.Node) string {
	var sb strings.Builder
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		if c.Type == html.TextNode {
			sb.WriteString(c.Data)
		}
	}
	return sb.String()
}

// resolve: ubah path aset relatif jadi URL absolut biar kebaca dari origin lain
func resolve(base *url.URL, href string) string {
	if href == "" {
		return ""
	}
	ref, err := url.Parse(href)
	if err != nil {
		return ""
	}
	abs := base.ResolveReference(ref)
	if abs.Scheme != "http" && abs.Scheme != "https" {
		return ""
	}
	return abs.String()
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
