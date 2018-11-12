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
	"fmt"
	"golang.org/x/sys/unix"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
	"syscall" // for Exec only
	"time"
)

var (
	exists               = []syscall.Errno{syscall.EEXIST}
	service_install_path = "/lib/systemd/system/"
	service_enable_path  = "/etc/systemd/system/multi-user.target.wants/"
	// Consider getting these from /dev/disk/by-partuuid/ or use root=/dev/mmcblk0p2
	partuuid_jessie  = "a8790229-02"
	partuuid_stretch = "4d3ee428-02"
	cmdline_template = "dwc_otg.lpm_enable=0 console=serial0,115200 console=tty1 root=PARTUUID=%s rootfstype=ext4 elevator=deadline fsck.repair=yes rootwait quiet init=/usr/lib/raspi-config/init_resize.sh"
)

func debugShell() {
	// This does not currently work. Gives the following error
	// bash: cannot set terminal process group (-1) inappropriate ioctl for device
	fmt.Printf("Starting a bash shell...")
	cmd := exec.Command("/bin/bash", "-i")
	cmd.Stdout = os.Stdout
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	fmt.Printf("Command finished with error: %v", err)

	fmt.Printf("Rebooting in 10 seconds...")
	time.Sleep(10 * time.Second)
	reboot()
}

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

func create_file(filename string, permissions os.FileMode, contents string) error {
	err := ioutil.WriteFile(filename, []byte(strings.TrimLeft(contents, "\r\n\t ")), permissions)
	if err != nil {
		return err
	}
	return nil
}

func create_service(name, contents string) error {
	src := service_install_path + name + ".service"
	dst := service_enable_path + name + ".service"
	err := create_file(src, 0644, contents)
	if err != nil {
		return err
	}
	err = os.Symlink(src, dst)
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
		unix.Mount("/", "/", "vfat", syscall.MS_REMOUNT, ""))
}

func mount_tmp() {
	checkFatalAllowed(
		"making tmp",
		unix.Mkdir("tmp", 0770),
		exists)
	checkFatal(
		"mounting tmp",
		unix.Mount("", "tmp", "tmpfs", 0, ""))
}

func mount_root() {
	checkFatalAllowed(
		"making new_root",
		unix.Mkdir("new_root", 0770),
		exists)
	checkFatal(
		"create device node",
		unix.Mknod("tmp/mmcblk0p2", 0660|syscall.S_IFBLK, 179<<8|2))
	checkFatal(
		"mounting real root",
		unix.Mount("tmp/mmcblk0p2", "new_root", "ext4", 0, ""))
}

func adjust_mounts() {
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

func replace_cmdline() {
	checkFatal(
		"renaming cmdline.txt to cmdline.txt.pi-init2",
		unix.Rename("/boot/cmdline.txt", "/boot/cmdline.txt.pi-init2"))

	b, err := ioutil.ReadFile("/usr/lib/os-release")
	checkFatal(
		"reading /usr/lib/os-release",
		err)

	str := string(b) // convert content bytes to a 'string'
	if strings.Contains(str, "(jessie)") {
		fmt.Println("it's jessie")
		checkFatal(
			"writing new cmdline.txt",
			create_file("cmdline.txt", 0644, fmt.Sprintf(cmdline_template, partuuid_jessie)))
	}
	if strings.Contains(str, "(stretch)") {
		fmt.Println("it's stretch")
		checkFatal(
			"writing new cmdline.txt",
			create_file("cmdline.txt", 0644, fmt.Sprintf(cmdline_template, partuuid_stretch)))
	}
}

func reboot() {
	unix.Sync()
	unix.Reboot(unix.LINUX_REBOOT_CMD_RESTART)
}

func customize() {
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
#!/bin/bash -eux

# Prevent *.sh from returning itself if there are no matches
shopt -s nullglob

# DISCLAIMER: I don't love this reboot solution. It was challenging to allow the user to use nested scripts and intercept the reboot from anywhere.
# A wrapper allow 'reboot' to be a valid completion of a task
# To use the bypass the wrapper and use the real executable, use /sbin/reboot (but even that is a symlink to /bin/systemctl)
tmp=$(mktemp --tmpdir -d pi-init2-XXX)
chmod a+rx $tmp
cat > $tmp/reboot <<'EOF'
#!/bin/bash -eux
  mv $(tail -n1 $(dirname $BASH_SOURCE)/runlist) /boot/run-once.d/completed/
  /sbin/halt --reboot
EOF
chmod a+x $tmp/reboot
export PATH=$tmp:$PATH

# write a log
date="$(date +%F@%H.%M.%S)"
logfile="/boot/run-once.d/completed/logfile-${date}.txt"
touch $logfile
exec > >(tee -i $logfile)
exec 2>&1

# Allow lazily named scripts to work
for script in /boot/run-once* $(run-parts --test /boot/run-once.d); do
    if [[ -f $script ]]; then
        echo "run-parts: executing $script"
        incdir=/boot/run-once.d/incomplete
        mkdir -p $incdir
        inc=$incdir/$(basename $script)
        mv $script $inc
        echo $inc>>$tmp/runlist
        $inc
        status=$?
        if $(exit $status); then
            mv $inc /boot/run-once.d/completed/
            rm -d $incdir || true
	else
	    echo "WARNING: $script exited with status $status"
        fi
    fi
done

# Run every on-boot script
run-parts -v /boot/on-boot.d
`)
	create_service("pi-init2", `
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
	//debugShell()
	customize()
	replace_cmdline()
	reboot()
}
