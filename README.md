V3 - just grpc +buf cli

https://connect.build/docs/go/getting-started

go install github.com/bufbuild/buf/cmd/buf@latest
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install github.com/bufbuild/connect-go/cmd/protoc-gen-connect-go@latest

buf lint
buf generate

docker build -t mtr.devops.telekom.de/maximilian_schubert/canary-bot:latest .
docker push mtr.devops.telekom.de/maximilian_schubert/canary-bot:latest

# Analyse Memory
https://go.dev/blog/pprof
https://pkg.go.dev/net/http/pprof

go tool pprof http://localhost:6060/debug/pprof/profile   # 30-second CPU profile
go tool pprof http://localhost:6060/debug/pprof/heap      # heap profile
go tool pprof http://localhost:6060/debug/pprof/block     # goroutine blocking profile

go tool pprof -http=:8082 http://localhost:6060/debug/pprof/profile
?seconds=30

# Remote Debug
dlv debug --listen=:8088 --headless