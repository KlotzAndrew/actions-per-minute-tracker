run:
	go run main.go

fixlines:
	sed -i 's/\r//g' $(file)
