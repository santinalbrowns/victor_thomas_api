up:
	`migrate -database "mysql://root:pass1234@/fashioncollection" -path ./migrations up`

down:
	`migrate -database "mysql://root:pass1234@/fashioncollection" -path ./migrations up`

build:
	@go build -o bin/api cmd/main.go

test:
	@go test -v ./...

run: build
	@./bin/api
