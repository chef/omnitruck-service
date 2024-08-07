FROM golang:1.22.5-alpine AS stage1

WORKDIR /app

RUN apk add make

COPY ./ ./ 
COPY omnitruck.yml.example omnitruck.yml

RUN go mod download 

RUN make build

EXPOSE 3000
EXPOSE 3001
EXPOSE 3002 

FROM golang:1.22.5-alpine 

COPY --from=stage1 /app/bin/omnitruck-service bin/omnitruck-service
COPY --from=stage1 /app/templates/ ../templates/
COPY --from=stage1 /app/omnitruck.yml omnitruck.yml

CMD bin/omnitruck-service start --config omnitruck.yml
