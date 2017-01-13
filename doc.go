// Package mux is a simple-mux, however, it provides a few nice features such as easy use of middleware and a router
// which doesn't automatically look at only prefixes (like the Go built-in mux).
//
// Middleware can be added to both prefixes and to endpoints, which means that the simple composability of this gives a
// lot of power. Inclusion of a middleware pipeline within the router also provides lots of power, plus less end-user
// code.
//
//     r := mux.New()
//
// This package has been intentionally kept simple at around 250 lines of code.
//
// Inspired by a few other routers and middlewares such as:
//
//     • https://github.com/justinas/alice
//     • https://golang.org/pkg/net/http/#ServeMux
//     • https://github.com/gorilla/mux
//     • https://github.com/julienschmidt/httprouter
//     • https://github.com/pressly/chi
//     • https://github.com/dimfeld/httptreemux
//     • https://github.com/bmizerany/pat
//     • https://github.com/pilu/traffic
//     • https://github.com/rcrowley/go-tigertonic
//     • https://github.com/mikespook/possum
//     • https://github.com/zenazn/goji/
//     • https://github.com/gocraft/web
//
package mux
