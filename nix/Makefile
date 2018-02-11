.PHONY: all clean

all:
	nix-build --cores 0 --max-jobs 3 --show-trace release.nix

clean:
	rm -f result*
