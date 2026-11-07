FROM golang:1.27-alpine AS build

WORKDIR /src

COPY pkg/ pkg/
COPY services/api/go.mod services/api/go.sum services/api/

WORKDIR /src/services/api
RUN go mod download

COPY services/api/ ./

RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/main ./cmd

FROM alpine:3.20.1 AS prod
WORKDIR /app
COPY --from=build /out/main /app/main
USER nobody
CMD ["./main"]
