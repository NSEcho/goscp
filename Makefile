all:
	mkdir build
	go build -o build/gotftp cmd/gotftp/main.go
	go build -o build/gosheller cmd/gosheller/main.go

clean:
	rm -rf build/
