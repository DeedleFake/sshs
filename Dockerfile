FROM scratch

MAINTAINER DeedleFake <deedlefake at hotmail dot com>

VOLUME ["/data"]
EXPOSE 8080

COPY sshs /sshs

ENTRYPOINT ["/sshs", "-root", "/data"]
