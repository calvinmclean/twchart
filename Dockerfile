FROM golang:1.24-alpine AS build
RUN mkdir /build
ADD . /build
WORKDIR /build
RUN go build -o twchart ./cmd/twchart/main.go

FROM alpine:latest AS production

RUN apk add --no-cache ca-certificates

RUN mkdir /app
WORKDIR /app
COPY --from=build /build/twchart .
ENTRYPOINT ["/app/twchart"]
