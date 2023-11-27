FROM golang:1.19.13-alpine3.18

WORKDIR /app

COPY . .

RUN apk --no-cache add make build-base

RUN go mod download

RUN make build

ARG API_PORT=8080
EXPOSE ${API_PORT}

CMD ["sh", "entrypoint.sh"]