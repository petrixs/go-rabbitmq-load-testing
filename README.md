#!/bin/bash
set -x
go mod init performance
go get -u github.com/rs/zerolog
go get -u github.com/wagslane/go-rabbitmq

go build -o /tmp/performance/go-rabbitmq-performance /tmp/performance/main.go

/tmp/performance/go-rabbitmq-performance -v=vhost -h=host -u=user -p=password -debug=true -e=exchange name -r=queue -t=100000 -s=1800 -m='{"message": true}'
