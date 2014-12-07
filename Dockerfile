FROM busybox:ubuntu-14.04

ADD https://tiego-artifacts.s3.amazonaws.com/teapot.tar.gz /teapot.tar.gz
RUN tar -zxf /teapot.tar.gz && \
    chmod +x /teapot && \
    rm /teapot.tar.gz
