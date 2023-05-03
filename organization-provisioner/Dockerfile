FROM golang:1.19-alpine AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /provisioner

FROM node:16-alpine

ENV PATH=$PATH:/root/.pulumi/bin

RUN apk update && \
    apk add --no-cache curl libc6-compat && \
    curl -fsSL https://get.pulumi.com/ | sh;

COPY --from=build /provisioner /provisioner

VOLUME /tmp/pulumi

EXPOSE 8000

CMD ["/provisioner"]
