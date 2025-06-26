FROM golang:1.23.6-alpine AS stage1

RUN mkdir /user && \
    echo 'nobody:x:65534:65534:nobody:/:' > /user/passwd && \
    echo 'nobody:x:65534:' > /user/group

WORKDIR /app

RUN apk add make

COPY ./ ./ 
COPY omnitruck.yml.example omnitruck.yml

RUN go mod download 

RUN make build

EXPOSE 3000
EXPOSE 3001
EXPOSE 3002 
USER nobody:nobody

FROM golang:1.23.6-alpine 

COPY --from=stage1 /app/bin/omnitruck-service bin/omnitruck-service
COPY --from=stage1 /app/templates/ templates/
COPY --from=stage1 /app/docs docs/
COPY --from=stage1 /app/omnitruck.yml omnitruck.yml

CMD bin/omnitruck-service start --config omnitruck.yml
