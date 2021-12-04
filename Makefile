SHELL := /bin/bash

proto:
	oto -template ./api/templates/server.go.plush \
		-out ./api/admin.gen.go \
		-ignore Ignorer \
		-pkg api \
		./api/admin
	gofmt -w ./api/admin.gen.go ./api/admin.gen.go
	oto -template ./api/templates/client.go.plush \
		-out ./pkg/rpc/client.gen.go \
		-ignore Ignorer \
		-pkg rpc \
		./api/admin
	gofmt -w ./pkg/rpc/client.gen.go ./pkg/rpc/client.gen.go
