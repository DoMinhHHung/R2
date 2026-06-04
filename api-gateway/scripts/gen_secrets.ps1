$secrets = @{}
Get-Content ".secrets" | ForEach-Object {
    if ($_ -match "^([^#][^=]*)=(.*)$") {
        $secrets[$matches[1].Trim()] = $matches[2].Trim()
    }
}

$jwtSecret    = $secrets["JWT_SECRET"]
$jwtIssuer    = $secrets["JWT_ISSUER"]
$redisPass    = $secrets["REDIS_PASSWORD"]

@"
APP_ENV=development
APP_PORT=8080
APP_TLS_CERT=
APP_TLS_KEY=

REDIS_ADDR=localhost:6379
REDIS_PASSWORD=$redisPass
REDIS_DB=0

JWT_SECRET=$jwtSecret
JWT_ISSUER=$jwtIssuer

RATE_LIMIT_IP=100
RATE_LIMIT_USER=200
RATE_LIMIT_API_KEY=500
RATE_LIMIT_WINDOW_SECONDS=60

OTEL_ENDPOINT=localhost:4317
OTEL_SERVICE_NAME=api-gateway

SERVICES_AUTH=http://localhost:8081
SERVICES_USER=http://localhost:8082
SERVICES_PROPERTY=http://localhost:8083
SERVICES_BOOKING=http://localhost:8084
SERVICES_PAYMENT=http://localhost:8085
SERVICES_NOTIFICATION=http://localhost:8086

CB_MAX_REQUESTS=5
CB_INTERVAL_SECONDS=10
CB_TIMEOUT_SECONDS=30
CB_FAILURE_RATIO=0.6

RETRY_MAX_ATTEMPTS=3
RETRY_WAIT_MIN_MS=100
RETRY_WAIT_MAX_MS=1000
"@ | Set-Content ".env"

Write-Host "✅ .env generated"

function ToBase64($str) {
    return [Convert]::ToBase64String([System.Text.Encoding]::UTF8.GetBytes($str))
}

$jwtSecretB64  = ToBase64 $jwtSecret
$jwtIssuerB64  = ToBase64 $jwtIssuer
$redisPassB64  = ToBase64 $redisPass

@"
apiVersion: v1
kind: Secret
metadata:
  name: gateway-secret
  namespace: rental
type: Opaque
data:
  JWT_SECRET: $jwtSecretB64
  JWT_ISSUER: $jwtIssuerB64
  REDIS_PASSWORD: $redisPassB64
"@ | Set-Content "deployments/k8s/secret.yaml"

Write-Host "✅ deployments/k8s/secret.yaml generated"
Write-Host ""
Write-Host "Summary:"
Write-Host "  JWT_ISSUER : $jwtIssuer"
Write-Host "  JWT_SECRET : $($jwtSecret.Substring(0, [Math]::Min(6, $jwtSecret.Length)))***"