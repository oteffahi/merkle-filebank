FROM golang:1.21 as build
WORKDIR /src
COPY ./ .
RUN go build -o /bin/filebankd filebankd/main.go

FROM ubuntu:latest
COPY --from=build /bin/filebankd /bin/
RUN chmod +x /bin/filebankd

RUN useradd -ms /bin/bash filebankd
USER filebankd
WORKDIR /home/filebankd

RUN filebankd init
COPY --from=build /src/certs/server/* /home/filebankd/.filebankd/cert/
COPY --from=build /src/certs/ca/* /home/filebankd/.filebankd/cert/

CMD ["bash"]
