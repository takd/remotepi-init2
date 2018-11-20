#!/bin/bash -eux

disk=/dev/sda
time sudo dd status=progress if=/dev/zero of=$disk bs=4k && sync
echo -e "o\nn\np\n1\n\n\nw\n" | sudo fdisk $disk
