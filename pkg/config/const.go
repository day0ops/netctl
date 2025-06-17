package config

const (
	DefaultQemuSystem                 = "qemu:///system"
	DefaultBridge                     = "virbr0"
	DefaultPrivateMinikubeNetworkName = "minikube-net"

	NetworkTmpl = `
<network>
  <name>{{.Name}}</name>
  <dns enable='no'/>
  <bridge name='{{.Bridge}}' stp='on' delay='0'/>
  {{- with .Parameters}}
  <ip address='{{.Gateway}}' netmask='{{.Netmask}}'>
    <dhcp>
      <range start='{{.ClientMin}}' end='{{.ClientMax}}'/>
    </dhcp>
  </ip>
  {{- end}}
</network>
`
)

const (
	AppName = "netctl"
)

// VersionInfo is the application version info
type VersionInfo struct {
	Version  string
	Revision string
}

func AppVersion() VersionInfo { return VersionInfo{Version: appVersion, Revision: revision} }

var (
	appVersion = "development"
	revision   = "unknown"
)
