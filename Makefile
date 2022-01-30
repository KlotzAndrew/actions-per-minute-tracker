run:
	go run main.go

build:
	go build .

fixlines:
	sed -i 's/\r//g' $(file)
