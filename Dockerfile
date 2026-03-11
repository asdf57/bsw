FROM golang:1.25-alpine AS prereqs

COPY go.mod go.sum ./
RUN go mod download

FROM prereqs AS final

WORKDIR /app

COPY . .

RUN go build -o bsw .

EXPOSE 8080

CMD ["./bsw"]
