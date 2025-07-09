# Go image base for build go app
FROM golang:1.23.3-alpine AS builder

# Ser working Direcetory on container
WORKDIR /app 

# Copy go.mod and go.sum files
COPY go.mod go.sum ./ 
RUN go mod download

# Copy Source code
COPY . . 

# build Go app
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags="-s -w" -o /app/main .

# Final stage
FROM alpine:latest

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/main .

# คัดลอกไฟล์ .env เข้าไปใน image
# เราจะสร้าง .env.prod สำหรับ production แยกต่างหาก
COPY .env.prod .env

# บอกให้ Container เปิด Port 8080 รอรับ request
EXPOSE 8080

# คำสั่งที่จะรันเมื่อ container เริ่มทำงาน
CMD ["/app/main"]