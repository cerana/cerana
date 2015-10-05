cflags=-I/usr/include/libzfs -I/usr/include/libspl -Wno-write-strings
libs=-lzfs -lzfs_core -lnvpair

all: nv

nv: nv.cc

%:%.cc
	g++ -Og -std=c++11 -ggdb3 -W -Wall -Wextra -o $@ $^ ${cflags} ${libs}

nv.txt: nv

.PHONY: known_good_data_test.go
known_good_data_test.go: ../known_good_data_test.go

../known_good_data_test.go: nv
	./$^ > $@

clean:
	rm -rf nv known_good_data_test.go
