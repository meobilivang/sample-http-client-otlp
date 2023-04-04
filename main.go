package main

import (
	"context"
	"fmt"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

/*
The application consists of two services; serviceA and serviceB. Customers send requests to serviceA, which in turn calls serviceB.
When called, serviceB adds two numbers and then returns the result as part of the SVC-RESPONSE http header. ServiceA echos back that header to the customer/client.
*/

const serviceName = "AdderSvc"

func main() {
	ctx := context.Background()

	// Start tracing in the app
	{
		tracerProvider, err := setupTracing(ctx, serviceName)
		if err != nil {
			panic(err)
		}
		defer tracerProvider.Shutdown(ctx)
	}

	go serviceA(ctx, 8081)

	serviceB(ctx, 8082)
}

func serviceA(ctx context.Context, port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/svcA", serviceA_HttpHandler)

	// add http middleware -> add traces to http handlers
	handler := otelhttp.NewHandler(mux, "server.http")

	serverPort := fmt.Sprintf(":%d", port)
	server := &http.Server{Addr: serverPort, Handler: handler}

	fmt.Println("Service A listening on", server.Addr)

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

func serviceB(ctx context.Context, port int) {
	mux := http.NewServeMux()
	mux.HandleFunc("/svcB", serviceB_HttpHandler)

	// add http middleware -> add traces to http handlers
	handler := otelhttp.NewHandler(mux, "server.http")

	serverPort := fmt.Sprintf(":%d", port)
	server := &http.Server{Addr: serverPort, Handler: handler}

	fmt.Println("Service B listening on", server.Addr)

	if err := server.ListenAndServe(); err != nil {
		panic(err)
	}
}

func serviceA_HttpHandler(w http.ResponseWriter, r *http.Request) {
	// Create trace span
	ctx, span := otel.Tracer("myTracer").Start(r.Context(), "serviceA_HttpHandler")
	defer span.End()

	// http.RoundTripper -> add traces to http requests
	cli := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8082/svcB", nil)

	if err != nil {
		panic(err)
	}

	resp, err := cli.Do(req)
	if err != nil {
		panic(err)
	}
	w.Header().Add("SVC-RESPONSE", resp.Header.Get("SVC-RESPONSE"))
}

func serviceB_HttpHandler(w http.ResponseWriter, r *http.Request) {
	// create trace span
	ctx, span := otel.Tracer("myTracer").Start(r.Context(), "serviceB_HttpHandler")
	defer span.End()

	answer := add(ctx, 42, 1813)
	w.Header().Add("SVC-RESPONSE", fmt.Sprint(answer))

	fmt.Fprintf(w, "Hi from ServiceB! The answer is: %d", answer)
}

func add(ctx context.Context, x, y int64) int64 {
	// create trace span
	_, span := otel.Tracer("myTracer").Start(
		ctx,
		"add",
		// add labels/tags/resources(if any) that are specific to this scope.
		trace.WithAttributes(attribute.String("component", "addition")),
		trace.WithAttributes(attribute.String("someKey", "someValue")),
		trace.WithAttributes(attribute.Int("age", 89)),
	)
	defer span.End()

	return x + y
}
