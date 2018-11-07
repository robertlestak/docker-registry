FROM golang:1.11

RUN apt-get update -y && apt-get install -y apache2-utils \
    apt-transport-https \
    ca-certificates \
    curl \
    software-properties-common \
    apt-transport-https dirmngr

RUN echo 'deb https://apt.dockerproject.org/repo debian-stretch main' >> /etc/apt/sources.list && \
    apt-key adv --keyserver hkp://p80.pool.sks-keyservers.net:80 --recv-keys F76221572C52609D && \
    apt-get update -y && apt-get install -y docker-engine

WORKDIR /src

COPY . .

RUN go build -o registrymanager .

EXPOSE 8081

CMD ["./registrymanager"]
