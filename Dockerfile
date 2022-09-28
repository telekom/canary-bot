##
## Build
##
FROM dockerhub.devops.telekom.de/golang:1.18-alpine AS build

ADD . /app

WORKDIR /app

RUN go mod download && \
    go build -o ./app/app-build .

##
## Deploy
##
FROM dockerhub.devops.telekom.de/alpine:3.13

WORKDIR /

ARG user_id=1001
RUN adduser -S $user_id -G root -u $user_id

COPY --from=build --chown=$user_id:root ./app/app ./app

USER $user_id

ENTRYPOINT ["/app/app-build"]