.PHONY: all clean update local shell

all:
	nix-build --cores 0 --max-jobs 3 default.nix

clean:
	rm -f result*

shell:
	nix-shell --pure

update:
	pkgs/os-specific/linux/cerana/update.sh

local:
	pkgs/os-specific/linux/cerana/update-local.sh
