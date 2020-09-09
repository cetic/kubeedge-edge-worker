FROM golang:alpine as dev
RUN apk add git
RUN apk add build-base
RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN go get github.com/eclipse/paho.mqtt.golang
RUN go get github.com/goombaio/namegenerator
RUN go get github.com/kubeedge/kubeedge/edge/pkg/devicetwin/dttype
RUN go get gopkg.in/yaml.v2
RUN go build -o main .
CMD sh

FROM alpine
COPY --from=dev /app/main ./app/main
CMD ./app/main
