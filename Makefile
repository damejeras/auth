SHELL := /bin/bash

spec:
	oto -template ./api/templates/server.go.plush \
		-out ./api/admin.gen.go \
		-ignore Ignorer \
		-pkg api \
		./api/admin
	gofmt -w ./api/admin.gen.go ./api/admin.gen.go
	oto -template ./api/templates/client.go.plush \
		-out ./api/client/client.gen.go \
		-ignore Ignorer \
		-pkg client \
		./api/admin
	gofmt -w ./api/client/client.gen.go ./api/client/client.gen.go
