# netctl - Network Command Line Tool for libvirt

This project helps to create the underlying network for kvm/qemu using `libvirt` go [bindings](https://pkg.go.dev/libvirt.org/go/libvirt#section-readme).
Only supported on Linux platforms.

The reason this exists is higher level tools such as `virsh` often require root privileges to create networks similarly.