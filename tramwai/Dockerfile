FROM golang:1.24.3
WORKDIR /tr
COPY tramwai/ ./tramwai
COPY scrapper/ ./scrapper
WORKDIR /tr/tramwai
RUN go mod download
RUN go build -o tramwai .
RUN apt update && apt install python3 python3-bs4 python3-requests -y
WORKDIR /tr
CMD ["/tr/tramwai/tramwai"]
