# Distributed Tracing with Jaeger

This backend uses OpenTelemetry with Jaeger for distributed tracing and observability.

## Overview

Jaeger is integrated to provide:
- **Distributed Tracing**: Track requests across the entire application
- **Performance Monitoring**: Identify bottlenecks and slow operations
- **Request Flow Visualization**: See how requests flow through the system
- **Error Tracking**: Identify and debug issues quickly

## Architecture

The tracing setup uses:
- **OpenTelemetry SDK**: Industry-standard observability framework
- **Jaeger Exporter**: Sends traces to Jaeger collector
- **Gorilla Mux Middleware**: Automatically instruments HTTP requests
- **Automatic Span Creation**: Each HTTP request creates a trace span

## Accessing Jaeger UI

Once the application is running with Docker Compose:

1. **Start the services:**
   ```bash
   docker-compose up -d
   ```

2. **Access Jaeger UI:**
   - URL: http://localhost:16686
   - Open in your browser to view traces

3. **View Traces:**
   - Select service: `splitwise-backend`
   - Click "Find Traces"
   - View detailed trace information

## Configuration

### Environment Variables

- `JAEGER_ENDPOINT`: Jaeger collector endpoint (default: `http://jaeger:14268/api/traces`)

### Docker Compose

The Jaeger service is automatically configured in `docker-compose.yml`:

```yaml
jaeger:
  image: jaegertracing/all-in-one:latest
  ports:
    - "16686:16686"  # Jaeger UI
    - "14268:14268"  # Jaeger collector HTTP
```

## How It Works

1. **Request Arrives**: HTTP request hits the backend
2. **Middleware Creates Span**: Tracing middleware creates a trace span
3. **Span Propagation**: Span context is propagated through the request
4. **Trace Export**: Completed spans are sent to Jaeger collector
5. **Visualization**: Traces appear in Jaeger UI

## Trace Information

Each trace includes:
- **Service Name**: `splitwise-backend`
- **HTTP Method**: GET, POST, etc.
- **HTTP Path**: `/users`, `/expenses`, etc.
- **Status Code**: 200, 404, 500, etc.
- **Duration**: Request processing time
- **Timestamps**: Start and end times

## Example Traces

### Creating a User
```
POST /users
├── Request parsing
├── Service call: CreateUser
└── Response encoding
```

### Getting Expenses by Group
```
GET /groups/{group_id}/expenses
├── Parameter extraction
├── Service call: GetExpensesByGroup
└── Response encoding
```

## Troubleshooting

### Traces Not Appearing

1. **Check Jaeger is running:**
   ```bash
   docker-compose ps jaeger
   ```

2. **Check backend logs:**
   ```bash
   docker-compose logs backend
   ```
   Look for: "Jaeger tracing initialized successfully"

3. **Verify endpoint:**
   ```bash
   curl http://localhost:16686/api/services
   ```

### Connection Issues

If the backend can't connect to Jaeger:
- Ensure Jaeger service is started before backend
- Check network connectivity: `docker-compose exec backend ping jaeger`
- Verify `JAEGER_ENDPOINT` environment variable

## Local Development

For local development without Docker:

1. **Start Jaeger locally:**
   ```bash
   docker run -d --name jaeger \
     -p 16686:16686 \
     -p 14268:14268 \
     jaegertracing/all-in-one:latest
   ```

2. **Set environment variable:**
   ```bash
   export JAEGER_ENDPOINT=http://localhost:14268/api/traces
   ```

3. **Run backend:**
   ```bash
   cd backend
   go run main.go
   ```

## Production Considerations

For production deployments:

1. **Sampling**: Adjust sampling rate to reduce overhead
2. **Batch Export**: Traces are batched for efficiency
3. **Resource Limits**: Configure appropriate resource limits
4. **Security**: Secure Jaeger endpoints
5. **Retention**: Configure trace retention policies

## Additional Resources

- [Jaeger Documentation](https://www.jaegertracing.io/docs/)
- [OpenTelemetry Go](https://opentelemetry.io/docs/instrumentation/go/)
- [Gorilla Mux Instrumentation](https://pkg.go.dev/go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux)

