## Todo API in Go

this repo contains source code for simple task management api with login and register, following best practices for api development in go.

## Database choice

We have chosen PostgreSQL as our database for persistent storage due to its adherence to ACID compliance and its robust support for concurrency.

## Setup

This service is using docker and docker-compose. the repo contains `docker-compose.yml` file which contains two docker service one is our task api written using go and another service for postgres database. so to run both service, please ensure you have docker installed in your system.

copy and configure `.env`

```
cp .env.template .env
```

Running both service
```
docker-compose up --build
```

you may run database only for local development, in that case use below command
```
docker-compose up dbhost
```

Additionaly we have `Makefile` for better tooling and development.

Use below command to build application binary
```
make build
```

You may also below command to run application on host machine without docker, in that case you will need to run the database container only.

```
make run
```

## API docs

The `docs` folder includes a Postman collection documented with various examples.