{ stdenv, cerana, utillinux, coreutils, systemd, gnugrep, gawk, zfs, bash, gptfdisk, grub2, lshw }:

stdenv.mkDerivation {
  name = "cerana-scripts-${cerana.rev}";

  src = cerana.src;

  installPhase = ''
    make DESTDIR=$out -C $src/boot install
    substituteInPlace $out/bin/fixterm --replace "stty" "${coreutils}/bin/stty"
    substituteInPlace $out/scripts/gen-hostid.sh --replace "tr" "${coreutils}/bin/tr"
    substituteInPlace $out/scripts/gen-hostid.sh --replace "uuidgen" "${utillinux}/bin/uuidgen"
  '';

  meta = with stdenv.lib; {
    homepage    = http://cerana.org;
    description = "Cerana OS";
    license     = licenses.mit ;
    platforms   = platforms.linux;
  };
}
