package port

import "net/http"

type ProxyRouter interface {
	Route(method, path string) (target string, ok bool)
	ServeHTTP(target string, w http.ResponseWriter, r *http.Request) error
}
