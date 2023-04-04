package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"os"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.12.0"
	"google.golang.org/grpc/credentials"
)

func setupTracing(ctx context.Context, serviceName string) (*trace.TracerProvider, error) {
	c, err := getTls()
	if err != nil {
		return nil, err
	}

	// Creates an exporter backed by gPRC
	//  -> creates trace data in OTLP wire format
	// 	-> export metrics to a OTEL collector
	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint("localhost:4317"), // connect to COLLECTOR
		otlptracegrpc.WithTLSCredentials(
			credentials.NewTLS(c),
		),
	)

	if err != nil {
		return nil, err
	}

	// labels/tags/res common to all traces
	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(serviceName),
		attribute.String("some-attribute", "some-value"),
	)

	provider := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource),
		// sampling rate based on parent span = 60%
		trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(0.6))),
	)

	// TracerProvider -> factory for Tracer
	// init Tracer + Exporter + Resource
	otel.SetTracerProvider(provider)

	// setting a context propagation option
	otel.SetTextMapPropagator(
		propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
		),
	)

	return provider, nil
}

func getTls() (*tls.Config, error) {
	clientAuth, err := tls.LoadX509KeyPair("./confs/client.crt", "./confs/client.key")

	if err != nil {
		return nil, err
	}

	caCert, err := os.ReadFile("./confs/rootCA.crt")

	if err != nil {
		return nil, err
	}

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	c := &tls.Config{
		RootCAs:            caCertPool,
		Certificates:       []tls.Certificate{clientAuth},
		InsecureSkipVerify: true,
	}

	return c, nil
}
