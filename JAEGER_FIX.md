# Jaeger Tracing Fix Summary

## Issue Fixed
Jaeger UI wasn't showing backend traces due to a resource initialization error.

## Problem
The backend was failing to initialize Jaeger tracing with the error:
```
cannot merge resource due to conflicting Schema URL
```

## Solution
Fixed the resource creation in `backend/adapter/inbound/http/tracing.go`:
- Changed from `resource.Merge()` to `resource.New()` 
- This avoids schema URL conflicts when creating the OpenTelemetry resource

## Verification

### Check Backend Tracing Status
```bash
docker-compose logs backend | grep -i "jaeger\|tracing"
```
Should show: `Jaeger tracing initialized successfully`

### Check Services in Jaeger
```bash
curl http://localhost:16686/api/services
```
Should include: `"splitwise-backend"`

### View Traces in Jaeger UI
1. Open http://localhost:16686 in your browser
2. Select service: `splitwise-backend`
3. Click "Find Traces"
4. You should see traces for all API requests

### Generate Test Traces
```bash
# Make some API calls
curl -X POST http://localhost:3000/api/users \
  -H "Content-Type: application/json" \
  -d '{"name":"Test User","email":"test@example.com"}'

curl http://localhost:3000/api/users
```

Wait a few seconds, then check Jaeger UI for new traces.

## Frontend Tracing

**Note**: The frontend is a React app running in the browser. Currently, only backend traces are instrumented. 

To add frontend tracing, you would need to:
1. Install OpenTelemetry JavaScript packages
2. Instrument the React app
3. Configure it to send traces to Jaeger

However, the backend traces already show the complete request flow from the frontend through Nginx to the backend, including:
- HTTP method and path
- Request duration
- Status codes
- Client IP addresses

## Current Status

✅ **Backend tracing**: Working
- Service name: `splitwise-backend`
- All HTTP requests are traced
- Traces appear in Jaeger UI

⚠️ **Frontend tracing**: Not implemented
- Would require OpenTelemetry JS instrumentation
- Backend traces show the full request flow

## Troubleshooting

If traces still don't appear:

1. **Check backend logs:**
   ```bash
   docker-compose logs backend
   ```
   Should show: "Jaeger tracing initialized successfully"

2. **Verify Jaeger is running:**
   ```bash
   docker-compose ps jaeger
   ```

3. **Check network connectivity:**
   ```bash
   docker-compose exec backend ping jaeger
   ```

4. **Verify endpoint:**
   ```bash
   docker-compose exec backend wget -O- http://jaeger:14268/api/traces
   ```

5. **Check service registration:**
   ```bash
   curl http://localhost:16686/api/services
   ```

6. **Make API calls and wait:**
   - Traces are batched and sent periodically
   - Wait 5-10 seconds after making requests
   - Refresh Jaeger UI

## Configuration

Backend tracing is configured via environment variable:
- `JAEGER_ENDPOINT`: Default is `http://jaeger:14268/api/traces`

This is set in `docker-compose.yml`:
```yaml
environment:
  - JAEGER_ENDPOINT=http://jaeger:14268/api/traces
```

