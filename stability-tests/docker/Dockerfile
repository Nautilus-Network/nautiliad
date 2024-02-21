ARG KASPAD_IMAGE
ARG KASPAMINER_IMAGE

FROM ${KASPAD_IMAGE} as nautiliad
FROM ${KASPAMINER_IMAGE} as nautilusminer

FROM golang:1.19-alpine

RUN mkdir -p /go/src/github.com/Nautilus-Network/nautiliad/

WORKDIR /go/src/github.com/Nautilus-Network/nautiliad/

RUN apk add bash build-base git

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

COPY --from=nautiliad /app/ /app/
COPY --from=nautilusminer /app/ /app/
ENV PATH="/app:${PATH}"

WORKDIR /go/src/github.com/Nautilus-Network/nautiliad/stability-tests

RUN go install ./...

ENTRYPOINT ["./run/run.sh"]