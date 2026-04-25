FROM golang:1.26-trixie AS build
RUN mkdir /build
ADD . /build
WORKDIR /build
ENV CGO_ENABLED=1
RUN go build -o twchart ./cmd/twchart/main.go

FROM debian:trixie-slim AS production

RUN apt-get update && apt-get install -y \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN mkdir /app
WORKDIR /app
COPY --from=build /build/twchart .
ENTRYPOINT ["/app/twchart"]
