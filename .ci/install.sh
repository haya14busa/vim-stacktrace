#!/bin/bash
set -ex

# Install Vim
root=$(cd $(dirname $0); pwd)
bash $root/installer/${TRAVIS_OS_NAME}.sh

