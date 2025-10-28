go mod tidy
docker build -t go-svc:local .
docker run -p 8080:8080 go-svc:local

# In another terminal
curl -s localhost:8080/ | jq
curl -s localhost:8080/healthz | jq
curl -s localhost:8080/status | jq
open http://localhost:8080/ui
open http://localhost:8080/swagger/index.html
