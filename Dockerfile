FROM dockerhub.devops.telekom.de/golang:1.18-alpine

ADD . /app

WORKDIR /app

RUN go mod download

RUN ls -lisa
RUN go build -o ./app/app-build .
RUN cd ./app && ls -lisa


ARG user_id=1001
RUN adduser -S $user_id -G root -u $user_id

RUN ls -lisa /app

#RUN setcap cap_net_raw=+ep ./app/app-build

USER $user_id

ENTRYPOINT ["/app/app-build"]