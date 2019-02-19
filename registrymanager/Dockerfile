FROM golang:1.11

RUN apt-get update -y && apt-get install -y apache2-utils \
    apt-transport-https \
    ca-certificates \
    curl \
    software-properties-common \
    apt-transport-https dirmngr

RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add - && \
    add-apt-repository \
   "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
   xenial \
   stable" && \
   apt-get update && \
   apt-get install -y docker-ce docker-ce-cli containerd.io


WORKDIR /src

COPY . .

RUN go build -o registrymanager .

EXPOSE 8081

CMD ["./registrymanager"]
