FROM golang:1.24.3
WORKDIR /app
COPY . .
RUN go mod download
ENV HTTP_PORT=42069
EXPOSE ${ENV}
RUN go build -o backend .
CMD ["./backend"]
