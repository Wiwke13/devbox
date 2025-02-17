#!/bin/bash

set -euo pipefail

declare -r pkg   # package that we're patching
declare -r glibc # new glibc that we're patching in
declare -r out   # nix output path that will contain the patched package

# Paths to this script's dependencies set by nix.
declare -r coreutils file findutils patchelf ripgrep

# Explicitly declare the specific commands that this script depends on.
hash -p "$coreutils/bin/cp" cp
hash -p "$coreutils/bin/chmod" chmod
hash -p "$coreutils/bin/cut" cut
hash -p "$coreutils/bin/echo" echo
hash -p "$coreutils/bin/head" head
hash -p "$coreutils/bin/mktemp" mktemp
hash -p "$coreutils/bin/rm" rm
hash -p "$coreutils/bin/stat" stat
hash -p "$coreutils/bin/wc" wc
hash -p "$file/bin/file" file
hash -p "$findutils/bin/find" find
hash -p "$findutils/bin/xargs" xargs
hash -p "$patchelf/bin/patchelf" patchelf
hash -p "$ripgrep/bin/rg" rg

# Copy the contents of the original package so we can patch them.
cp -R "$pkg" "$out"

# Find the new linker that we'll patch into all of the package's executables as
# the interpreter.
interp="$(find "$glibc/lib" -type f -maxdepth 1 -executable -name 'ld-linux-*.so*' | head -n1)"
readonly interp

patch() {
	declare -r binary="$1" # ELF binary to patch

	perm=$(stat -c "%a" "$binary")
	old_rpath="$(patchelf --print-rpath "$binary")"
	new_rpath="$glibc/lib${old_rpath:+:$old_rpath}"

	echo "running patchelf file=\"$binary\" rpath=\"$new_rpath\" perm=\"$perm\""
	chmod u+w "$binary"
	patchelf --set-rpath "$new_rpath" \
	         --add-needed libBrokenLocale.so.1 \
	         --add-needed libanl.so.1 \
	         --add-needed libc.so.6 \
	         --add-needed libdl.so.2 \
	         --add-needed libgcc_s.so.1 \
	         --add-needed libm.so.6 \
	         --add-needed libmvec.so.1 \
	         --add-needed libnsl.so.1 \
	         --add-needed libnss_compat.so.2 \
	         --add-needed libnss_db.so.2 \
	         --add-needed libnss_dns.so.2 \
	         --add-needed libnss_files.so.2 \
	         --add-needed libnss_hesiod.so.2 \
	         --add-needed libpcprofile.so \
	         --add-needed libpthread.so.0 \
	         --add-needed libresolv.so.2 \
	         --add-needed librt.so.1 \
	         --add-needed libutil.so.1 \
	         --set-interpreter "$interp" \
                 "$binary"

	# Neaten the runpath by removing extraneous paths. This will likely remove any old glibc.
	patchelf --shrink-rpath "$binary"
	chmod "$perm" "$binary"
}

# Search for any files that look like ELF binaries and patch them.
elves="$(find "$out" -type f -exec "$file/bin/file" {} \+ | rg --replace '$1' '^(.*): .*ELF.*executable.*dynamically linked.*$')"
count="$(echo "$elves" | wc -l)"
echo "patching elf binaries count=$count"
for binary in $elves; do
	patch "$binary" exe
done
