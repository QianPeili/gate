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

func (h *Host) Handle(w http.ResponseWriter, r *http.Request) {
	if handler, _ := h.m.Match(r.RequestURI); handler != nil {
		handler.(http.Handler).ServeHTTP(w, r)
		return
	}

	http.NotFound(w, r)
}

type Gate struct {
	m *mux.Mux
}

func NewGate() *Gate {
	m := mux.New()
	m.Matcher = mux.SuffixMatch
	return &Gate{
		m: m,
	}
}

func (g *Gate) Map(pattern string, h *Host) {
	g.m.Map(pattern, h)
}

func (g *Gate) Delete(pattern string) {
	g.m.Delete(pattern)
}

func (g *Gate) Handle(w http.ResponseWriter, r *http.Request) {
	if host, _ := g.m.Match(r.Host); host != nil {
		host.(*Host).Handle(w, r)
		return
	}

	http.NotFound(w, r)
}
