FROM golang:1.8.0

## Create a directory and Add Code
RUN mkdir -p /go/src/github.com/catpie/ss-go-mu
WORKDIR /go/src/github.com/catpie/ss-go-mu
ADD .  /go/src/github.com/catpie/ss-go-mu

# Download and install any required third party dependencies into the container.
RUN go-wrapper download
RUN go-wrapper install

EXPOSE 10000-20000

# Now tell Docker what command to run when the container starts
CMD ["go-wrapper", "run"]