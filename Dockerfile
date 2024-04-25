FROM golang:1.22.2 as go-builder
WORKDIR /app
COPY go.* ./
COPY *.go ./
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /testinbox

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=go-builder /testinbox /app/testinbox
COPY views/ /app/views/
COPY public/ /app/public/
ENV PORT 8080
EXPOSE 8080
CMD ["/app/testinbox"]
