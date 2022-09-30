FROM dockerhub.devops.telekom.de/alpine:3.16
ADD ./build /app
WORKDIR /app
ARG user_id=1001
RUN adduser -S $user_id -G root -u $user_id \
  && chown -R $user_id:root /app

USER $user_id

ENTRYPOINT ["/app/cbot"]