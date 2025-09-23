FROM golang:1.24-alpine AS build
RUN mkdir /build
ADD . /build
WORKDIR /build
RUN go build -o twchart . && chmod +x twchart

FROM alpine:latest AS production
RUN mkdir /app
WORKDIR /app
COPY --from=build /build/twchart .
ENTRYPOINT ["/app/twchart"]
