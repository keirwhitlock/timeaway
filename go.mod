module github.com/rorycl/timeaway

go 1.21

replace github.com/rorycl/timeaway/trips => ./trips

replace github.com/rorycl/timeaway/web => ./web

replace github.com/rorycl/timeaway/cmd => ./cmd

require (
	github.com/braintree/manners v0.0.0-20160418043613-82a8879fc5fd
	github.com/gorilla/handlers v1.5.1
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/schema v1.2.0
	github.com/jessevdk/go-flags v1.5.0
// github.com/rorycl/timeaway/trips v0.0.0
)

require (
	github.com/felixge/httpsnoop v1.0.3 // indirect
	golang.org/x/sys v0.12.0 // indirect
)
