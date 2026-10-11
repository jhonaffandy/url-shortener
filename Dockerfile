# FROM <image>:<tag> -> base image. Tag versi yg dipake
FROM golang:1.27

# WORKDIR <path> -> folder kerja di dalam container (dibuat otomatis)
WORKDIR /app

# COPY <src di hotst> <dest di container>
# go.mod & go.sum dicoyp duluan biar layer download dependency ke-cache
COPY go.mod go.sum ./

# RUN <command> -> jalan saat BUILD image
RUN go mod download

# Copy sisa source code
COPY . .

# go build -o <nama output> <package yg dibuild>
# "." = package di folder saat ini (WORKDIR /app)
RUN go build -o url-shortener .

# EXPOSE <port> -> dokumentasi doang, belum buka port ke host
EXPOSE 8080

# CMD ["executable","arg1", ...] -> jalan saat CONTAINER start
CMD ["./url-shortener"]