# Raspberry Pi Pre Init

## Purpose

A program which lets you set up a Raspberry Pi solely by writing to the /boot partition 
 (i.e. the one you can write from most computers!).

This allows you to distribute a small .zip file to set up a Raspberry Pi to do anything.
 You tell the user to unzip it over the top of the Pi's boot partition - 
 the system can set itself up perfectly on the first boot.

Additionally, once a Raspberry Pi has been set up using [pi-init2](src/pi-init2/init.go),
 files under the `appliance` base directory are symlinked back to the /boot,
 allowing you to reliably edit those "user-serviceable" files from the computer in future.
 So e.g. the list of wireless networks and passwords,
 or other files specific to the kind of appliance you're building.

## Trying it out

- Download and write a standard [Raspbian Jessie SD card](https://www.raspberrypi.org/downloads/raspbian/),
  e.g. the [Raspbian Jessie Lite](https://downloads.raspberrypi.org/raspbian_lite_latest).
- Unzip the latest release into the /boot partition
- Remove the SD card and put it into your Pi.

The Raspberry Pi should now boot several times.
 The first boot takes 2-5 minutes depending on your network,
 and which model of Raspberry Pi you use (I tested with model 3).

By default the following changes will be applied:

- SSH will be enabled by adding the `/boot/ssh` file.
- The hostname will be set to the content of `/boot/hostname`.
- If GitHub is reachable, SSH keys will be downloaded and saved in `/home/pi/.ssh/authorized_keys`,
  and password authentication will be disabled.

**Beware**: You'll need to edit the `pi-install` script to use _your_ GitHub username (or remove that part completely)!

# Building pi-init2

You'll find a script called '/build-and-copy' which you can use from a Linux or MacOS
 to build the [pi-init2](src/pi-init2/init.go) program,
 copy all the appliance files into place,
 and unmount the card.

# Disclaimer/Credits

Credits go to the following projects:

- [pi-init2](https://github.com/BytemarkHosting/pi-init2): This project is actually a fork of pi-init2, but heavily modified and stripped down to my needs.
  That's why the binary is still named `pi-init2`, so that its origin won't be forgotten.
- [raspbian-boot-setup](https://github.com/RichardBronosky/raspbian-boot-setup): Another project with a similar technique.
- [PiBakery](https://github.com/davidferguson/pibakery): A good resource to find more blocks to setup your Raspberry Pi.

Any contributions appreciated!
