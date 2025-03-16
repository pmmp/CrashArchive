FROM golang:1.23.0 as build
WORKDIR /go/crasharchive
COPY . .
RUN go build -o ./bin/crasharchive ./cmd/crasharchive.go
RUN go build -o ./bin/crasharchive-adduser ./cmd/crasharchive-adduser.go

FROM debian:8.8

WORKDIR /app
COPY --from=build /go/crasharchive/bin/* ./

CMD ["./crasharchive"]
