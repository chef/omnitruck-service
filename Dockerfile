FROM golang:1.18-alpine

WORKDIR /app

RUN apk add make

COPY ./ ./ 
COPY omnitruck.yml.example omnitruck.yml

RUN go mod download 

RUN make all

EXPOSE 3000
EXPOSE 3001
EXPOSE 3002 

CMD bin/omnitruck-service start
