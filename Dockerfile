FROM alpine:latest
COPY cmd/finca/finca /usr/local/bin/finca
EXPOSE 8080
CMD ["/usr/local/bin/finca", "-h"]
