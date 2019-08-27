#!/bin/bash

##  Usage: This script will run once after systemd brings up the network and
#   then get moved into /boot/run-once.d/completed which will be created for
#   you. This script is in this project to serve as an demo of how you might use
#   it. There are several distinct tasks in commented blocks. The recommended
#   use would be to create the /boot/run-once.d/ directory yourself and put
#   each task in its own file and name them so they sort in the order you want
#   them ran.
#   See: http://manpages.ubuntu.com/manpages/bionic/man8/run-parts.8.html


#### Update hostname
##  See https://raspberrypi.stackexchange.com/a/66939/8375 for a list of all the raspi-config magic you may want ot automate.
raspi-config nonint do_hostname "$(cat /boot/hostname)"

### Update locale, keyboard, WiFi country
raspi-config nonint do_change_locale "hu_HU.UTF-8"
raspi-config nonint do_configure_keyboard "hu"
raspi-config nonint do_wifi_country "HU"

### Enable UART for status LED
raspi-config nonint do_onewire 1

#### Wifi Setup (WPA Supplicant)
##  Replaces the magic of https://github.com/RPi-Distro/raspberrypi-net-mods/blob/master/debian/raspberrypi-net-mods.service
##  See: https://www.raspberrypi.org/documentation/configuration/wireless/wireless-cli.md
cat /etc/wpa_supplicant/wpa_supplicant.conf /boot/network.conf > /etc/wpa_supplicant/wpa_supplicant.conf
chmod 600 /etc/wpa_supplicant/wpa_supplicant.conf
wpa_cli -i wlan0 reconfigure

#### SSH Daemon Setup
##  Replaces the magic of https://github.com/RPi-Distro/raspberrypi-sys-mods/blob/master/debian/raspberrypi-sys-mods.sshswitch.service
##  See also: https://github.com/RPi-Distro/raspberrypi-sys-mods/blob/master/debian/raspberrypi-sys-mods.regenerate_ssh_host_keys.service
update-rc.d ssh enable && invoke-rc.d ssh start
dd if=/dev/hwrng of=/dev/urandom count=1 bs=4096
rm -f -v /etc/ssh/ssh_host_*_key*
/usr/bin/ssh-keygen -A -v

#### Setup own services
mv /boot/services /home/pi/services
chmod +x /home/pi/services/DS1302/setDateToRTC.py
chmod +x /home/pi/services/DS1302/getDateFromRTC.py

### Setup RTC time load at startup
echo "$(sed '$ i\/home/pi/services/DS1302/getDateFromRTC.py' /etc/rc.local)" > /etc/rc.local
### Run it too
/home/pi/services/DS1302/getDateFromRTC.py

### Copy shutdown button and wifi indicator services
## http://www.diegoacuna.me/how-to-run-a-script-as-a-service-in-raspberry-pi-raspbian-jessie/
cp /home/pi/shutdown-button.service /lib/systemd/system/shutdown-button.service
chmod 644 /lib/systemd/system/shutdown-button.service
cp /home/pi/wifi-checker.service /lib/systemd/system/wifi-checker.service
chmod 644 /lib/systemd/system/wifi-checker.service

### Enable the new services
sudo systemctl daemon-reload
sudo systemctl enable shutdown-button.service
sudo systemctl start shutdown-button.service
sudo systemctl enable wifi-checker.service
sudo systemctl start wifi-checker.service

#### Get additional scripts for subsequent usage, get will be run manually
mv /boot/get-remotepi-scripts /home/pi/get-remotepi-scripts
chmod +x get-remotepi-scripts
