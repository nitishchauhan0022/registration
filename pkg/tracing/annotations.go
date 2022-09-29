package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
)

const TracePrefix string = "trace.ocm.io/"

// Trace carrier for carring annotations satisfying the TextMapCarrier interface.
type TraceCarrier map[string]string

// Get returns the value associated with the passed key.
func (o TraceCarrier) Get(key string) string {
	return o[TracePrefix+key]
}

// Set stores the key-value pair.
func (o TraceCarrier) Set(key string, value string) {
	o[TracePrefix+key] = value
}

// Keys lists the keys stored in this carrier.
func (o TraceCarrier) Keys() []string {
	keys := make([]string, 0, len(o))
	for k := range o {
		keys = append(keys, k)
	}
	return keys
}

// Injects the trace to the annotations.
func TraceToAnnotation(ctx context.Context, annotations map[string]string) {
	otel.GetTextMapPropagator().Inject(ctx, TraceCarrier(annotations))
}

// Adding trace to the annotation of the object or updating the annotation
// if there is a trace present already
func TraceToObject(ctx context.Context, obj runtime.Object) error {
	m, err := meta.Accessor(obj)
	if err != nil {
		return err
	}
	annotations := m.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	TraceToAnnotation(ctx, annotations)
	m.SetAnnotations(annotations)
	return nil
}


// Builds a span from annotations if trace found, otherwise a new from the context
func SpanFromAnnotations(ctx context.Context, name string, spanName string, annotations map[string]string) (context.Context, trace.Span) {
	ctx = otel.GetTextMapPropagator().Extract(ctx, TraceCarrier(annotations))
	ctx, sp := otel.Tracer(name).Start(ctx, spanName)
	return ctx, sp
}

// Builds span from the current object annotations if trace found
// otherwise a new span from context provided
func SpanFromObject(ctx context.Context, name string,spanName string, obj runtime.Object) (context.Context, trace.Span) {
	m, err := meta.Accessor(obj)
	if err != nil {
		return otel.Tracer(name).Start(ctx, spanName)
	}
	ctx, sp := SpanFromAnnotations(ctx, name, spanName, m.GetAnnotations())
	return ctx, sp
}
