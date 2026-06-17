FROM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY cmd/ cmd/
COPY internal/ internal/
RUN go build -o /go-nc-mcp-web ./cmd/server/

FROM alpine:3.21
RUN apk add --no-cache ca-certificates
COPY --from=build /go-nc-mcp-web /usr/local/bin/go-nc-mcp-web
EXPOSE 8080
ENTRYPOINT ["go-nc-mcp-web"]
