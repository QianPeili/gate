package gate

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/huangml/mux"
)

type Host struct {
	m *mux.Mux
}

func NewHost() *Host {
	return &Host{
		m: mux.NewPathMux(),
	}
}

func (h *Host) Map(pattern, destURL string) error {
	if !strings.Contains(destURL, "://") {
		destURL = "http://" + destURL
	}

	u, err := url.Parse(destURL)
	if err != nil {
		return err
	}

	h.m.Map(pattern, httputil.NewSingleHostReverseProxy(u))
	return nil
}

func (h *Host) Delete(pattern string) {
	h.m.Delete(pattern)
}

func (h *Host) Clear() {
	h.m.Clear()
}

func (h *Host) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if handler := h.m.Match(r.RequestURI); handler != nil {
		handler.(http.Handler).ServeHTTP(w, r)
		return
	}

	http.NotFound(w, r)
}

type Gate struct {
	m *mux.Mux
}

func NewGate() *Gate {
	return &Gate{
		m: mux.New(mux.Config{
			TrimPattern: hostTrim,
			Matcher:     hostMatch,
		}),
	}
}

func (g *Gate) Map(pattern string, h *Host) {
	g.m.Map(pattern, h)
}

func (g *Gate) Delete(pattern string) {
	g.m.Delete(pattern)
}

func (g *Gate) Clear() {
	g.m.Clear()
}

func (g *Gate) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if host := g.m.Match(r.Host); host != nil {
		host.(*Host).ServeHTTP(w, r)
		return
	}

	http.NotFound(w, r)
}

func hostTrim(s string) string {
	if n := strings.Index(s, "://"); n >= 0 {
		s = s[n+3:]
	}

	if n := strings.IndexAny(s, "/"); n >= 0 {
		s = s[:n]
	}

	return s
}

func hostMatch(pattern, s string, index int) (ok bool, score int) {
	n := len(pattern)

	// "*" matches all hosts
	if pattern == "*" {
		return true, n
	}

	// "example.com:*" matches all ports
	if strings.HasSuffix(pattern, ":*") {
		pattern = strings.TrimSuffix(pattern, ":*")
		if n := strings.LastIndex(s, ":"); n > 0 {
			s = s[:n]
		}
	}

	// *.example.com matches example.com or xyz.example.com
	if n > 2 && pattern[:2] == "*." {
		return s == pattern[2:] || strings.HasSuffix(s, pattern[1:]), n
	}

	return s == pattern, n
}
