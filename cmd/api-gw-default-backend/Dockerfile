FROM golang:alpine
RUN mkdir /build 
ADD . /build/
WORKDIR /build 
#RUN go build -o main .
RUN go mod download goji.io
RUN CGO_ENABLED=0 GOOS=linux GO111MODULE=on go build -a -ldflags '-extldflags "-static"' -o ./api-gw-default-backend ./main.go
#FROM alpine
#RUN adduser -S -D -H -h /app appuser
#USER appuser
#COPY --from=builder /build/main /app/
#WORKDIR /app
#EXPOSE 8000
#CMD ["./main"]
