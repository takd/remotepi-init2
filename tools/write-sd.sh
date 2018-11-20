#!/bin/bash -eux

ftype=zip
mnt="/media/$(id -un)"

find_image(){
  (
    cd $(dirname $BASH_SOURCE)/..
    find $PWD/images -type f -name "*.$ftype" | sort | head -n1
  )
}

find_device(){
  disks=($(ls /dev/disk/by-id/usb-Generic_* | awk '!/part[0-9]/{print}'))
  disk_count=${#disks[@]}
  if [[ $disk_count -ne 1 ]]; then
    echo "Cannot proceed unless a single disk is found. Found $disk_count."
    exit 255
  else
    echo "${disks[0]}"
  fi
}

write_from_zip(){
  zip_size=$(unzip -l "$img" | awk 'END{print $1}')
  unzip -p "$img" | pv -s $zip_size | sudo dd of="$dev" status=progress bs=4M conv=fsync
}

write_from_img(){
  img_size=$(stat -c %s "$img")
  dd if="$img"    | pv -s $img_size | sudo dd of="$dev" status=progress bs=4M conv=fsync
}

unmount_device(){
  mnt_point=($(for m in $(realpath $dev*); do mount | awk -v m="$m" '$1==m{print $3}'; done))
  sudo umount  ${mnt_point[@]} || true
  sudo rm -rfv ${mnt_point[@]} || true
}

mount_device(){
  partition="$1-part1"
  label=$( source /dev/stdin <<<"$(sudo blkid -o export $partition)"; echo $LABEL; )
  mnt_point="$mnt/$label"
  sudo mkdir -p "$mnt_point"
  sudo mount $partition "$mnt_point"
}

copy_files(){
  (
    cd $(dirname $BASH_SOURCE)/..
    sudo cp -vr boot/* /media/pi/boot/
  )
}

write(){
  if [[ $ftype = zip ]]; then
    write_from_zip
  elif [[ $ftype = img ]]; then
    write_from_img
  else
    echo "This script does not know how to process files of type $ftype";
  fi
}

main(){
  img="$(find_image)"
  dev="$(find_device)"
  unmount_device
  write
  mount_device $dev
  copy_files
  unmount_device
}

if [ "$0" = "$BASH_SOURCE" ]; then main; fi
