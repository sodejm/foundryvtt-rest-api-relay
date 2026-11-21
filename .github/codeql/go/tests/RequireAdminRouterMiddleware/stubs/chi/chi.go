// Minimal inert API surface for static query extraction; no HTTP server runs.
package chi

import "net/http"

type Router interface {
	http.Handler
	Use(...func(http.Handler) http.Handler)
	With(...func(http.Handler) http.Handler) Router
	Group(func(Router)) Router
	Route(string, func(Router)) Router
	Mount(string, http.Handler)
	Get(string, http.HandlerFunc)
	Method(string, string, http.Handler)
	MethodFunc(string, string, http.HandlerFunc)
	Post(string, http.HandlerFunc)
}
type Mux struct{}

func NewRouter() Router                                       { return &Mux{} }
func (*Mux) ServeHTTP(http.ResponseWriter, *http.Request)     {}
func (*Mux) Use(...func(http.Handler) http.Handler)           {}
func (m *Mux) With(...func(http.Handler) http.Handler) Router { return m }
func (m *Mux) Group(f func(Router)) Router                    { f(m); return m }
func (m *Mux) Route(_ string, f func(Router)) Router          { f(m); return m }
func (*Mux) Mount(string, http.Handler)                       {}
func (*Mux) Get(string, http.HandlerFunc)                     {}
func (*Mux) Post(string, http.HandlerFunc)                    {}

func (*Mux) Method(string, string, http.Handler)         {}
func (*Mux) MethodFunc(string, string, http.HandlerFunc) {}
