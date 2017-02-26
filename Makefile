test:
	go test

bench:
	cd benchmark && go test -bench . -benchmem
