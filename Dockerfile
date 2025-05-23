FROM alpine:latest

WORKDIR /app

# Copy the pre-built Go backend binary
COPY reminder-app/main .

# Copy the pre-built static frontend assets
COPY www/dist ./static

EXPOSE 8080

ENTRYPOINT ["./main"]