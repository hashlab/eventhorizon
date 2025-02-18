// Copyright (c) 2020 - The Event Horizon authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package tracing

import (
	"context"
	"encoding/json"
	"log"

	eh "github.com/looplab/eventhorizon"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// The string keys to marshal the context.
const (
	tracingSpanKeyStr = "eh_tracing_span"
)

func init() {
	eh.RegisterContextMarshaler(func(ctx context.Context, vals map[string]interface{}) {
		if span := opentracing.SpanFromContext(ctx); span != nil {
			tracer := opentracing.GlobalTracer()

			carrier := opentracing.TextMapCarrier{}
			if err := tracer.Inject(span.Context(), opentracing.TextMap, &carrier); err != nil {
				log.Printf("eventhorizon: could not inject tracing span: %s", err)

				return
			}

			js, err := json.Marshal(carrier)
			if err != nil {
				log.Printf("eventhorizon: could not marshal tracing span: %s", err)

				return
			}

			vals[tracingSpanKeyStr] = string(js)
		}
	})
	eh.RegisterContextUnmarshaler(func(ctx context.Context, vals map[string]interface{}) context.Context {
		if js, ok := vals[tracingSpanKeyStr].(string); ok {
			tracer := opentracing.GlobalTracer()

			carrier := opentracing.TextMapCarrier{}
			if err := json.Unmarshal([]byte(js), &carrier); err != nil {
				log.Printf("eventhorizon: could not unmarshal tracing span: %s", err)

				return ctx
			}

			parentSpanContext, err := tracer.Extract(opentracing.TextMap, carrier)
			if err != nil && err != opentracing.ErrSpanContextNotFound {
				log.Printf("eventhorizon: could not extract tracing span: %s", err)

				return ctx
			}

			span := tracer.StartSpan("eventbus", ext.RPCServerOption(parentSpanContext))
			ctx = opentracing.ContextWithSpan(ctx, span)
		}

		return ctx
	})
}
