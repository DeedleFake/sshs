FROM scratch

MAINTAINER DeedleFake <deedlefake@users.noreply.github.com>

VOLUME ["/data"]
EXPOSE 8080

COPY sshs /sshs

ENTRYPOINT ["/sshs"]
