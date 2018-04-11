/* pi-init2
 *
 * A shim to drop onto a Raspberry Pi to write some files to its root
 * filesystem before giving way to the real /sbin/init.  Its goal is simply
 * to allow you to customise a RPi by dropping files into that FAT32 /boot
 * partition, as opposed to either 1) booting it and manually setting it up, or
 * 2) having to mount the root partition, which Windows & Mac users can't easily
 * do.
 *
 * Cross-compile on Mac/Linux:
 *   GOOS=linux GOARCH=arm go get golang.org/x/sys/unix
 *   GOOS=linux GOARCH=arm go build pi-init2
 *
 * Cross-compile:
 *   set GOOS=linux
 *   set GOARCH=arm
 *   go build packages pi-init2
 */

package main

import (
	"golang.org/x/sys/unix"
	"fmt"
	"io"
	"os"
	"io/ioutil"
	"strings"
	"syscall" // for Exec only
	"time"
)

var (
	exists = []syscall.Errno{syscall.EEXIST}
	service_install_path = "/lib/systemd/system/"
	service_enable_path  = "/etc/systemd/system/multi-user.target.wants/"
)

func checkFatalAllowed(desc string, err error, allowedErrnos []syscall.Errno) {
	if err != nil {
		errno, ok := err.(syscall.Errno)
		if ok {
			for _, b := range allowedErrnos {
				if b == errno {
					return
				}
			}
		}
		fmt.Println("error " + desc + ":" + err.Error())
		time.Sleep(10 * time.Second)
		unix.Exit(1)
	}
}

func checkFatal(desc string, err error) {
	checkFatalAllowed(desc, err, []syscall.Errno{})
}

// from https://gist.github.com/elazarl/5507969
func cp(dst, src string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	// no need to check errors on read only file, we already got everything
	// we need from the filesystem, so nothing can go wrong now.
	defer s.Close()

	d, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(d, s); err != nil {
		d.Close()
		return err
	}
	return d.Close()
}

func create_file(filename string, permissions os.FileMode, contents string) {
	ioutil.WriteFile(filename, []byte(strings.TrimLeft(contents, "\r\n\t ")), permissions)
}

func create_service(name, contents string) error {
	src := service_install_path + name + ".service"
	dst := service_enable_path  + name + ".service"
	create_file(src, 0644, contents)
	err := os.Symlink(src, dst)
	if err != nil {
		return err
	}
	return nil
}

func remount_rw() {
	checkFatal(
		"changing directory",
		unix.Chdir("/"))
	checkFatal(
		"remount rw",
		unix.Mount("/", "/", "vfat", syscall.MS_REMOUNT, ""), )
}

func mount_tmp(){
	checkFatalAllowed(
		"making tmp",
		unix.Mkdir("tmp", 0770),
		exists)
	checkFatal(
		"mounting tmp",
		unix.Mount("", "tmp", "tmpfs", 0, ""))
}

func mount_root(){
	checkFatalAllowed(
		"making new_root",
		unix.Mkdir("new_root", 0770),
		exists)
	checkFatal(
		"create device node",
		unix.Mknod("tmp/mmcblk0p2", 0660 | syscall.S_IFBLK, 179<<8 | 2))
	checkFatal(
		"mounting real root",
		unix.Mount("tmp/mmcblk0p2", "new_root", "ext4", 0, ""))
}

func adjust_mounts(){
	checkFatal(
		"pivoting",
		unix.PivotRoot("new_root", "new_root/boot")) // new_root becomes root FS & current root FS moves to new_root/boot
		// See: https://linux.die.net/man/8/pivot_root
	checkFatal(
		"unmounting /boot/tmp",
		unix.Unmount("/boot/tmp", 0))
	checkFatal(
		"removing /boot/new_root",
		os.Remove("/boot/new_root"))
	checkFatal(
		"removing /boot/tmp",
		os.Remove("/boot/tmp"))
	checkFatal(
		"changing into boot directory",
		unix.Chdir("/boot"))
}

func replace_cmdline(){
	checkFatal(
		"renaming cmdline.txt to cmdline.txt.pi-init2",
		unix.Rename("/boot/cmdline.txt", "/boot/cmdline.txt.pi-init2"))
	checkFatal(
		"renaming cmdline.txt.orig to cmdline.txt",
		unix.Rename("/boot/cmdline.txt.orig", "/boot/cmdline.txt"))
}

func reboot(){
	unix.Sync()
	unix.Reboot(unix.LINUX_REBOOT_CMD_RESTART)
}

func customize(){
	checkFatal(
		"changing into boot directory",
		unix.Chdir("/boot"))
	checkFatalAllowed(
		"making on-boot.d",
		unix.Mkdir("on-boot.d", 0770),
		exists)
	checkFatalAllowed(
		"making run-once.d",
		unix.Mkdir("run-once.d", 0770),
		exists)
	checkFatalAllowed(
		"making run-once.d/completed",
		unix.Mkdir("run-once.d/completed", 0770),
		exists)
	create_file("/usr/local/sbin/pi-init2-run-parts.sh", 0744, `
#!/bin/bash

# Prevent *.sh from returning itself if there are no matches
shopt -s nullglob

# Allow lazily named scripts to work
for script in /boot/run-once*; do
	if [[ -f $script ]]; then
		$script
		status=$?
		if $(exit $status); then
			mv $script /boot/run-once.d/completed/
		fi
	fi
done

# Run every run-once script
run-parts --exit-on-error /boot/run-once.d 2>/tmp/completed
sed -i '/^run-parts: executing/!d;s/^run-parts: executing *//' /tmp/completed

# Pop last script off the list if run-parts exited on an error
status=$?
if ! $(exit $status); then
	sed -i '$d' /tmp/completed
fi

# Move completed scripts
while read script; do
	mv $script /boot/run-once.d/completed/
done < /tmp/completed

# Run every on-boot script
run-parts /boot/on-boot.d
`)
	create_service("pi-init2",`
[Unit]
Description=Run user provided scripts on boot
ConditionPathExists=/usr/local/sbin/pi-init2-run-parts.sh
After=network-online.target raspi-config.service

[Service]
ExecStart=/usr/local/sbin/pi-init2-run-parts.sh
Type=oneshot
TimeoutSec=600

[Install]
WantedBy=multi-user.target
`)
}

func main() {
	remount_rw()
	mount_tmp()
	mount_root()
	adjust_mounts()
	customize()
	replace_cmdline()
/*
	checkFatal(
		"exec real init",
		syscall.Exec("/sbin/init", os.Args, nil))
*/
	reboot()
}
