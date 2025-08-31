FROM golang:1.24-alpine

WORKDIR /app

COPY . ./

RUN go mod tidy
RUN go mod vendor

RUN go build -o /service

CMD [ "/service" ]