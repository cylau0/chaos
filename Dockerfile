FROM golang:1.17 as build-env

WORKDIR /go/src/app

COPY . .

RUN go mod tidy
#RUN go get download .
#RUN go build
#RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /app .
RUN CGO_ENABLED=0 go build -a -installsuffix cgo -o /app .


FROM alpine:latest  
RUN apk --no-cache add ca-certificates 
WORKDIR /
COPY --from=build-env /app /

CMD ["./app"]  
