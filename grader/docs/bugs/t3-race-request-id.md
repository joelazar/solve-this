# t3-race-request-id

Tier 3. Planted in `internal/api/middleware.go`, caught by `TestT3RaceRequestID`.

The request id counter loses its atomic type:

```go
var requestCounter uint64

func requestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCounter++
		id := "req-" + strconv.FormatUint(requestCounter, 10)
		w.Header().Set("X-Request-Id", id)
		next.ServeHTTP(w, r)
	})
}
```

The HTTP server runs one goroutine per connection, so `requestCounter++` is an
unsynchronized read-modify-write on shared state. Increments get lost under load and
two responses can format the same counter value, so they share an `X-Request-Id`.
Single requests always look fine.

The baseline uses `atomic.Uint64` and `requestCounter.Add(1)`.

The hidden test hammers `GET /health` on a `-race` build and fails when the server log
contains a data race report naming `requestID`. `go vet` stays silent on this one; the
race detector under load is the tool that finds it.
