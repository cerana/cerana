.PHONY: all clean

all:
	nix-build --cores 0 --max-jobs 3 default.nix

clean:
	rm -f result*
