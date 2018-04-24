FROM debian:buster-slim
RUN apt-get update && apt-get install -y --no-install-recommends curl ca-certificates bzip2
RUN  curl -A "Docker Build" -sSL -o /tmp/luxrender.tar.bz2 http://www.luxrender.net/release/luxrender/1.6/linux/64/lux-v1.6-x86_64-sse2-NoOpenCL.tar.bz2 && \
    tar jxf /tmp/luxrender.tar.bz2 --strip-components=1 -C /usr/local/bin/ && \
    rm -rf /tmp/lux*
EXPOSE 18018
COPY cmd/finca/finca /usr/local/bin/finca
CMD ["/usr/local/bin/finca", "-h"]
