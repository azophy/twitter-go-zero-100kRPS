# syntax=docker/dockerfile:1
# adapted from https://docs.docker.com/language/golang/build-images/#multi-stage-builds

# Build the application from source
FROM golang:1.21-alpine AS build-stage

WORKDIR /app
EXPOSE 8080

RUN apk update \
 && apk add --no-cache git wrk

#RUN go install go.k6.io/xk6/cmd/xk6@latest \
 #&& xk6 build --with github.com/szkiba/xk6-dashboard@latest
RUN wget https://github.com/grafana/xk6-dashboard/releases/download/v0.7.2/xk6-dashboard_v0.7.2_linux_amd64.tar.gz \
 && tar -xvf xk6*.tar.gz

# start null so we could control the container manually
CMD "tail -f /dev/null"
