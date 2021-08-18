package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	name = flag.String("name", "tiny-echo", "Service name to reply with - defaults to tiny-echo")
	port = flag.Int("port", 80, "Port for the container to listen on - defaults to 80")
)

func init() {
	flag.Parse()
}

type Echo struct {
	Name     string                 `json:"name"`
	Host     string                 `json:"host"`
	Hostname string                 `json:"hostname"`
	Request  map[string]interface{} `json:"request,omitempty"`
	Message  string                 `json:"message,omitempty"`
}

func handler(w http.ResponseWriter, r *http.Request) {

	var clientIp string
	if r.Header.Get("X-Forwarded-For") != "" {
		clientIp = r.Header.Get("X-Forwarded-For")
	} else {
		clientIp = r.RemoteAddr
	}

	echo := Echo{
		Name:     *name,
		Host:     r.Host,
		Hostname: os.Getenv("HOSTNAME"),
		Request: map[string]interface{}{
			"method":   r.Method,
			"protocol": r.Proto,
			"path":     r.RequestURI,
			"query":    r.URL.RawQuery,
			"headers":  r.Header,
			"clientIp": clientIp,
		},
	}

	res, err := json.Marshal(echo)
	if err != nil {
		log.Fatalf("error marshaling json: %v", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(res)

	defer log.Println(string(res))

	return
}

func jobHandler(w http.ResponseWriter, r *http.Request) {
	// random sleep between 1 and 4 seconds to simulate long running job
	rand.Seed(time.Now().UnixNano())
	min := 1
	max := 4
	n := rand.Intn(max-min+1) + min
	time.Sleep(time.Duration(n) * time.Second)

	echo := Echo{
		Name:     *name,
		Host:     r.Host,
		Hostname: os.Getenv("HOSTNAME"),
		Message:  fmt.Sprintf("job successfully processed in %d seconds", n),
	}

	res, err := json.Marshal(echo)
	if err != nil {
		log.Fatalf("error marshaling json: %v", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write(res)

	defer log.Println(string(res))
}

func CORSMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, HEAD")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Max-Age", "3600")
			w.WriteHeader(http.StatusNoContent)
		}
		next.ServeHTTP(w, r)
	})
}

func main() {

	mux := http.NewServeMux()
	mux.Handle("/", CORSMiddleware(http.TimeoutHandler(http.HandlerFunc(handler), 200*time.Millisecond, "")))
	mux.Handle("/job", CORSMiddleware(http.TimeoutHandler(http.HandlerFunc(jobHandler), 5000*time.Millisecond, "")))

	ctx := context.Background()
	server := &http.Server{
		Addr:    fmt.Sprintf(":%v", *port),
		Handler: mux,
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		signalType := <-ch
		signal.Stop(ch)
		log.Printf("%v received. Exiting...", signalType)
		server.Shutdown(ctx)
		os.Exit(0)
	}()

	log.Printf("Starting tiny-echo server on port %d...", *port)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Printf("server exited: %s", err)
		os.Exit(1)
	}
}
