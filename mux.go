package mux

import (
	"context"
	"log"
	"net/http"
	"path"
	"strings"
)

type key int

const valsIdKey key = 999

// A constructor for a piece of middleware. Some middleware use this constructor out of the box, so in most cases you
// can just pass somepackage.New
// type Middleware func(http.Handler) http.Handler

type MuxHandler struct {
	Method      string
	Path        string
	Segments    []string
	Length      int
	Middlewares []func(http.Handler) http.Handler
	Handler     http.Handler
}

type Router struct {
	muxHandlers []MuxHandler
}

// Make sure the Router conforms with the http.Handler interface.
var _ http.Handler = New()

// New returns a new initialized Router.  Nothing is automatic. You must do slash/non-slash redirection yourself.
func New() *Router {
	return &Router{}
}

// Get is a shortcut for router.use("GET", path, ...handlers)
func (r *Router) Get(path string, middlewares []func(http.Handler) http.Handler, handler http.Handler) {
	r.use("GET", path, middlewares, handler)
}

// Post is a shortcut for router.use("POST", path, ...middlewares)
func (r *Router) Post(path string, middlewares []func(http.Handler) http.Handler, handler http.Handler) {
	r.use("POST", path, middlewares, handler)
}

// Use is a shortcut for router.use("GET", path, ...middlewares). Unlike other methods such as Get, Post, Put, Patch, and
// Delete, Use actually matches the path as a prefix and not the entire path. (Though of course, the entire path also
// matches.)
//
// e.g. r.Use("/profile/", ...) matches the requests "/profile/", "/profile/settings", and "/profile/a/path/to/".
//
// Note however, r.Use("/profile/", ...) doesn't match "/profile" since it contains too many slashes. But
// r.Use("/profile", ...) does match "/profile/" and "/profile/...".
//
// Also note that if you Use("/s/"), then this can be used to handle all static files inside /s/.
func (r *Router) Use(path string, middleware func(http.Handler) http.Handler) {
	r.use("USE", path, []func(http.Handler) http.Handler{middleware}, nil)
}

// Handle registers a new request handle with the given path and method.
//
// The respective shortcuts (for GET, POST, PUT, PATCH and DELETE) can also be used.
func (r *Router) use(method, path string, middlewares []func(http.Handler) http.Handler, handler http.Handler) {
	if path[0] != '/' {
		panic("path must begin with '/' in path '" + path + "'")
	}

	if r.muxHandlers == nil {
		r.muxHandlers = make([]MuxHandler, 0)
	}

	segments := strings.Split(path, "/")[1:]
	muxHandler := MuxHandler{
		Method:      method,
		Path:        path,
		Segments:    segments,
		Length:      len(segments),
		Middlewares: middlewares,
		Handler:     handler,
	}

	log.Printf("muxHandler=%#v\n", muxHandler)

	// add it to the middlewares
	r.muxHandlers = append(r.muxHandlers, muxHandler)

	log.Printf("muxHandlers=%#v\n", r.muxHandlers)
}

func isPrefixMatch(segments []string, handler *MuxHandler) bool {
	log.Printf("isPrefixMatch: %v\n", segments)

	// can't match if the handler prefix length is longer than the URL
	if handler.Length > len(segments) {
		return false
	}

	// check each segment is the same (for the length of the prefix)
	for i, segment := range handler.Segments {
		log.Printf("isPrefixMatch: checking '%s' against '%s'\n", segments[i], segment)

		// if both segments are empty, then this matches
		if segment == "" && segments[i] == "" {
			log.Printf(" - both empty, fine\n")
			continue
		}

		// check if segment start with a ":"
		if segment[0:0] == ":" {
			log.Printf("Placeholder = %s\n", segment)
			continue
		}

		if segments[i] != segment {
			log.Printf(" - not the same, this prefix doesn't match\n")
			return false
		}
	}

	// nothing stopped us from matching, so it must be true
	return true
}

func isMatch(method string, segments []string, handler *MuxHandler) (map[string]string, bool) {
	log.Printf("isMatch: %s %v\n", method, segments)

	// can't match if the methods are different
	if handler.Method != method {
		log.Printf("isMatch: different method (got %s, this route is %s)\n", method, handler.Method)
		return nil, false
	}

	// can't match if the url length is different from the route length
	if handler.Length != len(segments) {
		log.Printf("isMatch: different path length (got %d, this route is %d long)\n", len(segments), handler.Length)
		return nil, false
	}

	vals := make(map[string]string)

	// check each segment is the same (for the length of the prefix)
	for i, segment := range handler.Segments {
		log.Printf("isMatch: checking '%s' against '%s'\n", segments[i], segment)

		// if both segments are empty, then this matches
		if segment == "" && segments[i] == "" {
			log.Printf(" - both empty, fine\n")
			continue
		}

		// check if segment start with a ":"
		if segment != "" && segment[0:1] == ":" {
			log.Printf("Placeholder = %s\n", segment)
			// ToDo: store/return this value somewhere
			vals[segment[1:]] = segments[i]
			continue
		}

		if segments[i] != segment {
			return nil, false
		}
	}

	// nothing stopped us from matching, so it must be true
	return vals, true
}

// ServeHTTP
func (router *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("--- NEW REQUEST %s %s ---\n", r.Method, r.URL.Path)

	method := r.Method
	normPath := path.Clean(r.URL.Path)
	log.Printf("request: method=%#v\n", method)
	log.Printf("request: path1=%#v\n", r.URL.Path)
	log.Printf("request: path2=%#v\n", normPath)

	// if the original path ends in a slash
	if normPath != "/" {
		if strings.HasSuffix(r.URL.Path, "/") {
			normPath = normPath + "/"
		}
	}

	// if these paths differ, then redirect to the real one
	if normPath != r.URL.Path {
		http.Redirect(w, r, normPath, http.StatusFound)
		return
	}

	log.Printf("request: split=%#v\n", strings.Split(normPath, "/"))
	segments := strings.Split(normPath, "/")[1:]

	log.Printf("request: segments=%#v\n", segments)

	for i, muxHandler := range router.muxHandlers {
		log.Printf("--- Route(%d): %s /%s\n", i, muxHandler.Method, strings.Join(muxHandler.Segments, "/"))

		var vals map[string]string
		var matched bool
		// check to see if this is just a USE (and therefore, a prefix match)
		if muxHandler.Method == "USE" {
			log.Printf("This route is a USE, therefore, we just check the prefix\n")
			matched = isPrefixMatch(segments, &muxHandler)
		} else {
			log.Printf("This route is a GET/POST/PUT/etc, therefore, we match the entire segment length\n")
			vals, matched = isMatch(method, segments, &muxHandler)
			log.Printf("vals1=%#v\n", vals)

			ctx := context.WithValue(r.Context(), valsIdKey, vals)
			r = r.WithContext(ctx)
		}

		// if matched, then we have a non-nil 'vals' too, even if it contains no values
		if matched {
			log.Printf("matched, calling all handlers in this route:")
			// loop over all handlers
			for j, middleware := range muxHandler.Middlewares {
				log.Printf(" - handler %d\n", j)

				// presume the middleware deals with this request fully (and doesn't call `next`)
				finished := true

				// call the middleware ... passing this `next` function to see if we want the next handler
				next := func(w http.ResponseWriter, r *http.Request) {
					log.Printf("*** next has been called\n")
					finished = false
				}

				nextHandlerFunc := http.HandlerFunc(next)

				// now call the middleware with our next
				log.Printf("before middleware\n")
				// log.Printf("vals2=%#v\n", vals)
				// ctx := context.WithValue(r.Context(), valsIdKey, vals)
				middleware(nextHandlerFunc).ServeHTTP(w, r)
				log.Printf("after middleware\n")

				// if we're finished, then just return
				if finished {
					return
				}
			}

			// finally, check if we have a handler and if so, call it - presume it is the last in the chain
			if muxHandler.Handler != nil {
				muxHandler.Handler.ServeHTTP(w, r)
				return
			}
		}
	}

	// if we got through to here, then either nothing matched, or any middlewares did match but still called `next` and
	// hence there is no final route to deal with the request
	http.NotFound(w, r)
}

func Vals(r *http.Request) map[string]string {
	return r.Context().Value(valsIdKey).(map[string]string)
}
