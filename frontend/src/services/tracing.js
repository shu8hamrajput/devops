import { trace } from '@opentelemetry/api';
import { WebTracerProvider } from '@opentelemetry/sdk-trace-web';
import { SemanticResourceAttributes } from '@opentelemetry/semantic-conventions';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';
import { FetchInstrumentation } from '@opentelemetry/instrumentation-fetch';
import { XMLHttpRequestInstrumentation } from '@opentelemetry/instrumentation-xml-http-request';
import { registerInstrumentations } from '@opentelemetry/instrumentation';
import { BatchSpanProcessor } from '@opentelemetry/sdk-trace-base';

/**
 * Initialize OpenTelemetry tracing for the frontend
 * This sets up distributed tracing with Jaeger
 */
export function initTracing() {
  try {
    // Get Jaeger endpoint from environment or use default
    // In production (Docker), use full URL with current origin (proxied through Nginx)
    // In development, use localhost directly
    let jaegerEndpoint = import.meta.env.VITE_JAEGER_ENDPOINT;
    
    if (!jaegerEndpoint) {
      if (import.meta.env.PROD) {
        // In production, construct full URL from current origin (nginx proxy)
        // Check if window is available (browser environment)
        if (typeof window !== 'undefined' && window.location) {
          jaegerEndpoint = `${window.location.origin}/v1/traces`;
        } else {
          // Fallback if window is not available
          console.warn('Window not available, skipping tracing initialization');
          return null;
        }
      } else {
        // In development, use direct Jaeger endpoint
        jaegerEndpoint = 'http://localhost:4318/v1/traces';
      }
    }

    // Create OTLP trace exporter
    const traceExporter = new OTLPTraceExporter({
      url: jaegerEndpoint,
      headers: {},
    });

    // Create span processor with exporter
    const spanProcessor = new BatchSpanProcessor(traceExporter);

    // Create tracer provider with span processor
    // Service name must be set using the exact semantic convention attribute key
    // Using both the constant and string literal to ensure compatibility
    const tracerProvider = new WebTracerProvider({
      resourceAttributes: {
        'service.name': 'splitwise-frontend',
        'service.version': '1.0.0',
        // Also set using semantic convention constant as fallback
        [SemanticResourceAttributes.SERVICE_NAME]: 'splitwise-frontend',
        [SemanticResourceAttributes.SERVICE_VERSION]: '1.0.0',
      },
      spanProcessors: [spanProcessor],
    });

    // Set global tracer provider
    trace.setGlobalTracerProvider(tracerProvider);

    // Register instrumentations
    registerInstrumentations({
      instrumentations: [
        new FetchInstrumentation({
          // Instrument fetch requests
          propagateTraceHeaderCorsUrls: [
            /.*/, // Allow all URLs to receive trace headers
          ],
          clearTimingResources: true,
        }),
        new XMLHttpRequestInstrumentation({
          // Instrument XMLHttpRequest (used by axios in some browsers)
          propagateTraceHeaderCorsUrls: [
            /.*/, // Allow all URLs to receive trace headers
          ],
          clearTimingResources: true,
        }),
      ],
    });

    console.log('OpenTelemetry tracing initialized for frontend');
    console.log('Jaeger endpoint:', jaegerEndpoint);

    // Return tracer provider for potential cleanup
    return tracerProvider;
  } catch (error) {
    console.error('Error initializing OpenTelemetry tracing:', error);
    // Return null to indicate initialization failed
    return null;
  }
}

