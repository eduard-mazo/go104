FROM scratch
COPY bin/go104-linux-ppc64le /go104
ENTRYPOINT ["/go104"]
