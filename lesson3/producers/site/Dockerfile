FROM golang:1.24-alpine

WORKDIR /app
ENV CGO_ENABLED=1
RUN apk add --no-cache build-base gcc musl-dev

COPY go.mod ./
COPY go.sum ./
COPY *go ./

RUN go mod tidy
RUN go mod vendor

RUN go build -a -tags musl -installsuffix cgo -ldflags '-extldflags "-static"' -o /service


CMD [ "/service" ]