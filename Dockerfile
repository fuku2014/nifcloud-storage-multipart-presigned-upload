FROM golang:1.12.1-alpine3.9 AS build-backend

ENV GO111MODULE on

WORKDIR /app
COPY . /app

RUN apk add --no-cache git ca-certificates && CGO_ENABLED=0 go build main.go

# ---

FROM node:10.15.3 AS build-frontend

WORKDIR /app
COPY ./frontend /app

RUN npm install -g yarn && yarn install && npm run build && npm run export

# ---

FROM scratch

WORKDIR /app
COPY --from=build-backend /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build-backend /app/main main
COPY --from=build-frontend /app/out frontend/out

EXPOSE  8080
CMD [ "/app/main"]
