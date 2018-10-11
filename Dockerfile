FROM golang:1.11.1-stretch AS build
COPY . .
RUN [ "./build" ]

