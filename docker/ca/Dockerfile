FROM golang:1.20.1 as build
WORKDIR /go/crasharchive
COPY ./go.mod ./go.mod
RUN go mod download
COPY . .
RUN go build -o ./bin/crasharchive ./cmd/crasharchive.go
RUN go build -o ./bin/crasharchive-adduser ./cmd/crasharchive-adduser.go

FROM debian:8.8

WORKDIR /app
COPY --from=build /go/crasharchive/bin/* ./

CMD ["./crasharchive"]
