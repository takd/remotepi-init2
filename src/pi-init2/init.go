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
 * Cross-compile on Windows:
 *   set GOOS=linux
 *   set GOARCH=arm 
 *   go build packages pi-init2
 */

package main

import "golang.org/x/sys/unix"
import "fmt"
import "io"
import "os"
import "path/filepath"
import "strings"
import "syscall" // for Exec only
import "time"

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

func copyAppliance(path string, info os.FileInfo, err error) error {
	info, err = os.Stat(path)
	if err != nil {
		// should only be called with real directories
		return err
	}

	// for now we don't care about permissions

	if info.IsDir() {
		if os.Mkdir("/"+path, os.FileMode(int(0755))) != nil {
			return err
		}
	} else {
		// remove any existing file in place, ignore error, but let's
		// not use RemoveAll to delete directories, not sure anything
		// useful can come of that
		os.Remove("/" + path)

		if err = cp("/"+path, "/boot/appliance_install/"+path); err != nil {
			return err
		}
		if strings.HasSuffix("/"+path, ".service") || strings.HasSuffix("/"+path, ".target") {
			os.Chmod("/"+path, os.FileMode(int(0644)))
		}

		fmt.Println("Moved " + path + " to /")
	}

	return nil
}

func symlinkAppliance(path string, info os.FileInfo, err error) error {
	info, err = os.Stat(path)
	if err != nil {
		// should only be called with real directories
		return err
	}

	// for now we don't care about permissions

	if info.IsDir() {
		if os.Mkdir("/"+path, os.FileMode(int(0755))) != nil {
			return err
		}
	} else {
		// remove any existing file in place, ignore error, but let's
		// not use RemoveAll to delete directories, not sure anything
		// useful can come of that
		os.Remove("/" + path)

		if os.Symlink("/boot/appliance/"+path, "/"+path) != nil {
			return err
		}

		fmt.Println("Symlinked " + path + " to /boot/appliance")
	}

	return nil
}

func main() {
	exists := []syscall.Errno{syscall.EEXIST}
	checkFatal("changing directory",
		unix.Chdir("/"))
	checkFatal("remount rw",
		unix.Mount("/", "/", "vfat", syscall.MS_REMOUNT, ""), )
	checkFatalAllowed(
		"making tmp",
		unix.Mkdir("tmp", 0770),
		exists)
	checkFatalAllowed(
		"making new_root", unix.Mkdir("new_root", 0770), exists)
	checkFatal("mounting tmp",
		unix.Mount("", "tmp", "tmpfs", 0, ""))
	checkFatal("create device node",
		unix.Mknod("tmp/mmcblk0p2", 0660|syscall.S_IFBLK, 179<<8|2))
	checkFatal("mounting real root",
		unix.Mount("tmp/mmcblk0p2", "new_root", "ext4", 0, ""))
	checkFatal("pivoting",
		unix.PivotRoot("new_root", "new_root/boot"))
	checkFatal("unmounting /boot/tmp",
		unix.Unmount("/boot/tmp", 0))
	checkFatal("removing /boot/new_root",
		os.Remove("/boot/new_root"))
	checkFatal("removing /boot/tmp",
		os.Remove("/boot/tmp"))
	checkFatal("changing into boot directory",
		unix.Chdir("/boot"))
	checkFatal("removing cmdline.txt",
		os.Remove("/boot/cmdline.txt"))
	checkFatal("renaming cmdline.txt.orig to cmdline.txt",
		unix.Rename("/boot/cmdline.txt.orig", "/boot/cmdline.txt"))
	checkFatal("changing into appliance_install directory",
		unix.Chdir("/boot/appliance_install"))
	checkFatal("copying appliance_install to root",
		filepath.Walk(".", copyAppliance))
	checkFatal("changing into appliance directory",
		unix.Chdir("/boot/appliance"))
	checkFatal("symlink appliance to root",
		filepath.Walk(".", symlinkAppliance))
	unix.Sync()
	unix.Reboot(unix.LINUX_REBOOT_CMD_RESTART)
}
