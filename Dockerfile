FROM cgr.dev/chainguard/go:latest as gobuild
WORKDIR /app
ADD . .
RUN go mod download
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o canary .

FROM scratch
COPY --from=gobuild /app/canary /canary
COPY --from=gobuild /etc/passwd /etc/passwd
USER 65532
ENTRYPOINT ["/canary"]