package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go/config"
	"github.com/uber/jaeger-client-go/zipkin"
	"github.com/uber/jaeger-client-go"
	"github.com/uber/jaeger-client-go/transport"
		"os"
	"time"
	"github.com/opentracing-contrib/go-stdlib/nethttp"
)

func main() {
	zipkinPropagator := zipkin.NewZipkinB3HTTPHeaderPropagator()
nethttp.ClientTrace(true)
	c, err := config.FromEnv()
	if err != nil {
		log.Fatal("Failed to create tracer configuration")
	}
	endp := os.Getenv("JAEGER_ENDPOINT")
	log.Printf("Using endpoint: %s\n", endp)
	t, _, err := c.NewTracer(
		config.Injector(opentracing.HTTPHeaders, zipkinPropagator),
		config.Extractor(opentracing.HTTPHeaders, zipkinPropagator),
		config.Reporter(jaeger.NewRemoteReporter(
			transport.NewHTTPTransport(endp),
			jaeger.ReporterOptions.BufferFlushInterval(time.Second))))

	if err != nil {
		log.Fatal("Could not create tracer: ", err)
	}

	http.HandleFunc("/", nethttp.MiddlewareFunc(t, rootHandler))
	http.HandleFunc("/chaining", nethttp.MiddlewareFunc(t, chainingHandler(t)))
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func rootHandler(rw http.ResponseWriter, r *http.Request) {
	fmt.Println(r.Header)
	fmt.Fprintf(rw, "Hello from go!")
}

func chainingHandler(tracer opentracing.Tracer) func(rw http.ResponseWriter, r *http.Request) {
	return func(rw http.ResponseWriter, r *http.Request) {

		s := tracer.StartSpan("get client wrapper /",
			opentracing.ChildOf(opentracing.SpanFromContext(r.Context()).Context()))

		req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/", "golang-app:8080"), nil)
		//req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("http://%s/", r.Host), nil)
		req = req.WithContext(r.Context())
		req, ht := nethttp.TraceRequest(tracer, req)
		defer ht.Finish()
		c := &http.Client{Transport: &nethttp.Transport{}}
		resp, err := c.Do(req)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		bodyBytes, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := resp.Body.Close(); err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		s.Finish()
		fmt.Fprintf(rw, "Chaining --> %s", string(bodyBytes))
	}
}
