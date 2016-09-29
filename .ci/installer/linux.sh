#!/bin/bash
tmp=$(mktemp -d)
url=https://github.com/vim/vim
git clone --depth 1 --single-branch $url $tmp
cd $tmp
exit
./configure --prefix="$HOME/vim" \
    --enable-fail-if-missing \
    --with-features=huge
make -j2
make install
