FROM golang:1.18 as builder

##
## Build
##
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN CGO_ENABLED=0 go build -v -o restreamer

##
## Build Panel
##
FROM node:18 as panelbuilder

WORKDIR /app

COPY web ./
RUN cd npm install -g yarn && yarn && yarn build

##
## Build
##
FROM alpine:3.14

WORKDIR /root/
COPY --from=panelbuilder /app/dist ./web/dist
COPY --from=builder /app/restreamer ./restreamer

EXPOSE 13232
VOLUME [ "/root/stream.json" ]

ENTRYPOINT ["./islog"]