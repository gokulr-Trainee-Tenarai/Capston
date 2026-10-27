$services = @(
  @{ID="auth-service"; Name="auth-service"; Address="localhost"; Port=8083},
  @{ID="product-service"; Name="product-service"; Address="localhost"; Port=8084},
  @{ID="order-service"; Name="order-service"; Address="localhost"; Port=8082}
)

foreach ($s in $services) {
  $body = $s | ConvertTo-Json
  Invoke-RestMethod -Method Put -Uri "http://localhost:8500/v1/agent/service/register" -ContentType "application/json" -Body $body
}

Write-Host "Services registered. Check http://localhost:8500"