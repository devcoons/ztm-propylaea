# builder image
FROM golang:1.18-alpine as builder
ENV GO111MODULE=on
ENV GOPROXY="https://proxy.golang.org,direct"
RUN mkdir /build
WORKDIR /build
COPY .app/ /build/

RUN go mod tidy; exit 0
RUN go mod tidy; exit 0
RUN go mod tidy; exit 0
RUN go get -u; exit 0;
RUN go mod tidy; exit 0
RUN go mod tidy; exit 0

RUN CGO_ENABLED=0 GOOS=linux go build -a -o application .

# generate clean, final image for end users
FROM alpine:latest
RUN apk add --update redis && \
    rm -rf /var/cache/apk/* && \
    mkdir /data && \
    chown -R redis:redis /data && \
    sed -i 's#logfile /var/log/redis/redis.log#logfile ""#i' /etc/redis.conf && \
    sed -i 's#daemonize yes#daemonize no#i' /etc/redis.conf && \
    sed -i 's#dir /var/lib/redis/#dir /data#i' /etc/redis.conf && \
    echo -e "# placeholder for local options\n" > /etc/redis-local.conf && \
    echo -e "include /etc/redis-local.conf\n" >> /etc/redis.conf
RUN mkdir /app
WORKDIR /app
COPY --from=builder /build/application .
RUN chmod +x application
# executable
EXPOSE 8080 8081
ENTRYPOINT [ "/bin/sh", "-c" , "(sleep 5 && ./application) & /usr/bin/redis-server --protected-mode no & sleep infinite;" ]
