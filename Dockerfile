FROM golang:1.26-alpine AS build

ENV CGO_ENABLED=0
ENV GOOS=linux

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o /otc-prometheus-exporter .

FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /otc-prometheus-exporter /otc-prometheus-exporter
EXPOSE 39100
ENTRYPOINT [ "/otc-prometheus-exporter" ]
