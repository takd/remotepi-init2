
# Raspberry Pi Pre Init *for Kubernetes!*

This branch will likely become a dedicated repo, but for now, I just needed to
get my code somewhere safe.

## Purpose

***See [the master branch](../..) of this project for a better explanation of
pi-init2.*** In this branch I will only address the specifics of how to build a
kubernetes cluster with the project.

This project attempts to make it possible for users of Mac, Windows, or Linux to
build a multi-node Kubernetes cluster of headleses RPis without changing a
single file. All you have to do is drop content onto the [FAT formatted] `/boot`
partion of a fresh Rasbian image. Yes, it is quite ambitious.

## Trying it out

- Download and write a standard [Raspbian SD card](https://www.raspberrypi.org/downloads/raspbian/),
  - e.g. the [Raspbian Stretch Lite](https://downloads.raspberrypi.org/raspbian_lite_latest).
  - A good GUI tool for writing the SD is [Etcher](https://etcher.io). (Though I
prefer to use `dd` at the CLI so I can script it all.)
- Copy the content of this project's [boot folder](boot) to the microSD card's
  /boot partition.
- Remove the SD card [safely], put it into your Pi [powered off], and the power it on.

The Raspberry Pi should now boot several times. The first boot just creates a
systemd service in the ext4 partion (which Mac and Windows users cannot touch).
The second boot is the same as vanilla Rasbian's first boot (which resizes the
ext4 partion to use the whole SD card. The third boot does an `apt upgrade` and
takes 2-5 minutes depending on your network, and which model of Raspberry Pi you
use (I tested with model 3). The other boots install Docker and Kubernetes,
update your `cmdline.txt` with the `cgroup` requirements, and elects a *master*
Kubernetes node for the others to join. This will take longer because Kubernetes
has a lot of requirements.

# How it works

This is really cool. The `cmdline.txt` specifies an `init=/pi-init2` kernel
argument to use a custom binary in this package in place of the usual systemd
init. That binary holds everything (including the ability to recreate the
original `cmdline.txt` file) needed to make Raspbian run the scripts in
`run-once.d` on subsequent boots.

## How/Why you should incorporate this project into your Raspberry Pi project

If you have a project you expect someone to run on an RPi (especially if it
would be the RPi's single purpose) you could provide your own `run-once.sh`
script that will clone your project, configure, and install it.

# Disclaimer/Credits

This has been tested with this (what I believe to be the latest release) version
of [Jessie](http://downloads.raspberrypi.org/raspbian_lite/images/raspbian_lite-2017-07-05/)
but the instructions above assume Stretch.

Credits go to the following projects:

- [pi-init2](https://github.com/gesellix/pi-init2): A project that was a fork of my raspbian-boot-setup project, that I then forked to make **this** project.
- [raspbian-boot-setup](https://github.com/RichardBronosky/raspbian-boot-setup): My first project attempting to accomplish the same goal.
- [PiBakery](https://github.com/davidferguson/pibakery): A good resource to find more blocks to setup your Raspberry Pi.

Any contributions appreciated!

