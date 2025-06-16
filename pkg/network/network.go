package network

import (
	"bytes"
	"fmt"
	"net"
	"text/template"
	"time"

	"github.com/pkg/errors"
	"libvirt.org/go/libvirt"

	"github.com/day0ops/netctl/pkg/config"
	"github.com/day0ops/netctl/pkg/log"
	"github.com/day0ops/netctl/pkg/util"
)

type Network struct {
	// The name of the network
	Name string

	// The name of the bridge to create
	Bridge string

	// Subnet of the network
	Subnet string

	// QEMU Connection URI
	ConnectionURI string
}

type libvirtNetwork struct {
	Name   string
	Bridge string
	Parameters
}

// EnsureNetwork is called to set up the network if one doesn't exist. If it does exist it will try to recreate it
func (n *Network) EnsureNetwork() error {
	conn, err := getConnection(n.ConnectionURI)
	if err != nil {
		return fmt.Errorf("failed opening libvirt connection: %w", err)
	}
	defer func() {
		if _, err := conn.Close(); err != nil {
			log.Errorf("failed closing libvirt connection: %v", lvErr(err))
		}
	}()

	log.Infof("ensuring network %s is active", n.Name)
	// retry once to recreate the network, but only if is not used
	if err := setupNetwork(conn, n.Name); err != nil {
		log.Debugf("network %s is inoperable, will try to recreate it: %v", n.Name, err)
		if err := n.DeleteNetwork(); err != nil {
			return errors.Wrapf(err, "deleting inoperable network %s", n.Name)
		}
		log.Debugf("deleted or skipped %s network", n.Name)
		if err := n.createNetwork(); err != nil {
			return errors.Wrapf(err, "recreating inoperable network %s", n.Name)
		}
		log.Debugf("🎉 successfully recreated %s network", n.Name)
		if err := setupNetwork(conn, n.Name); err != nil {
			return err
		}
		log.Debugf("🎉 successfully activated %s network", n.Name)
	}

	return nil
}

// createNetwork is not called directly. See EnsureNetwork
func (n *Network) createNetwork() error {
	if n.Name == config.DefaultPrivateMinikubeNetworkName {
		return fmt.Errorf("network can't be named %s. This is the name of the private network created by minikube by default", config.DefaultPrivateMinikubeNetworkName)
	}

	conn, err := getConnection(n.ConnectionURI)
	if err != nil {
		return fmt.Errorf("failed opening libvirt connection: %w", err)
	}
	defer func() {
		if _, err := conn.Close(); err != nil {
			log.Errorf("failed closing libvirt connection: %v", lvErr(err))
		}
	}()

	// Only create the network if it does not already exist
	if netp, err := conn.LookupNetworkByName(n.Name); err == nil {
		log.Warnf("found existing %s network, skipping creation", n.Name)

		if netXML, err := netp.GetXMLDesc(0); err != nil {
			log.Debugf("failed getting %s network XML: %v", n.Name, lvErr(err))
		} else {
			log.Debug(netXML)
		}

		if err := netp.Free(); err != nil {
			log.Errorf("failed freeing %s network: %v", n.Name, lvErr(err))
		}
		return nil
	}

	// retry up to 5 times to create kvm network
	for attempts, subnetAddr := 0, n.Subnet; attempts < 5; attempts++ {
		// rather than iterate through all the valid subnets, give up at 20 to avoid a lengthy user delay for something that is unlikely to work.
		// will be like 192.168.39.0/24,..., 192.168.248.0/24 (in increment steps of 11)
		var subnet *Parameters
		subnet, err = FreeSubnet(subnetAddr, 11, 20)
		if err != nil {
			log.Debugf("failed finding free subnet for private network %s after %d attempts: %v", n.Name, 20, err)
			return fmt.Errorf("un-retryable: %w", err)
		}

		// reserve last client ip address for multi-control-plane loadbalancer vip address in ha cluster
		clientMaxIP := net.ParseIP(subnet.ClientMax)
		clientMaxIP.To4()[3]--
		subnet.ClientMax = clientMaxIP.String()

		// create the XML for the private network from our networkTmpl
		tryNet := libvirtNetwork{
			Name:       n.Name,
			Bridge:     n.Bridge,
			Parameters: *subnet,
		}
		tmpl := template.Must(template.New("network").Parse(config.NetworkTmpl))
		var networkXML bytes.Buffer
		if err = tmpl.Execute(&networkXML, tryNet); err != nil {
			return fmt.Errorf("executing private network template: %w", err)
		}

		// define the network using our template
		log.Debugf("generated network template as XML:\n%s", networkXML.String())
		libvirtNet, err := conn.NetworkDefineXML(networkXML.String())
		if err != nil {
			return fmt.Errorf("defining network %s %s from xml %s: %w", n.Name, subnet.CIDR, networkXML.String(), err)
		}

		// and finally create & start it
		log.Debugf("creating network %s %s...", n.Name, subnet.CIDR)
		if err = libvirtNet.Create(); err == nil {
			log.Debugf("network %s %s created", n.Name, subnet.CIDR)
			if netXML, err := libvirtNet.GetXMLDesc(0); err != nil {
				log.Debugf("failed getting %s network XML: %v", n.Name, lvErr(err))
			} else {
				log.Debugf("dumping network information as XML:\n%s", netXML)
			}

			return nil
		}
		log.Debugf("failed creating network %s %s, will retry: %v", n.Name, subnet.CIDR, err)
		subnetAddr = subnet.IP
	}
	return fmt.Errorf("failed creating network %s: %w", n.Name, err)
}

func (n *Network) DeleteNetwork() error {
	conn, err := getConnection(n.ConnectionURI)
	if err != nil {
		return fmt.Errorf("failed opening libvirt connection: %w", err)
	}
	defer func() {
		if _, err := conn.Close(); err != nil {
			log.Errorf("failed closing libvirt connection: %v", lvErr(err))
		}
	}()

	log.Debugf("checking if network %s exists...", n.Name)
	libvirtNet, err := conn.LookupNetworkByName(n.Name)
	if err != nil {
		if lvErr(err).Code == libvirt.ERR_NO_NETWORK {
			log.Warnf("network %s does not exist. Skipping deletion", n.Name)
			return nil
		}
		return errors.Wrapf(err, "failed looking up network %s", n.Name)
	}
	defer func() {
		if libvirtNet == nil {
			log.Warnf("nil network, cannot free")
		} else if err := libvirtNet.Free(); err != nil {
			log.Errorf("failed freeing %s network: %v", n.Name, lvErr(err))
		}
	}()

	log.Debugf("network %s exists", n.Name)

	err = n.checkDomains(conn)
	if err != nil {
		return err
	}

	// when we reach this point, it means it is safe to delete the network

	log.Debugf("trying to delete network %s...", n.Name)
	deleteFunc := func() error {
		active, err := libvirtNet.IsActive()
		if err != nil {
			return err
		}
		if active {
			log.Debugf("destroying active network %s", n.Name)
			if err := libvirtNet.Destroy(); err != nil {
				return err
			}
		}
		log.Debugf("undefining inactive network %s", n.Name)
		return libvirtNet.Undefine()
	}
	if err := util.LocalRetry(deleteFunc, 10*time.Second); err != nil {
		return errors.Wrap(err, "deleting network")
	}
	log.Debugf("network %s deleted", n.Name)

	return nil
}

func setupNetwork(conn *libvirt.Connect, name string) error {
	n, err := conn.LookupNetworkByName(name)
	if err != nil {
		return fmt.Errorf("failed looking up network %s: %w", name, lvErr(err))
	}
	defer func() {
		if n == nil {
			log.Warn("nil network, cannot free")
		} else if err := n.Free(); err != nil {
			log.Errorf("failed freeing %s network: %v", name, lvErr(err))
		}
	}()

	// always ensure autostart is set on the network
	autostart, err := n.GetAutostart()
	if err != nil {
		return errors.Wrapf(err, "checking network %s autostart", name)
	}
	if !autostart {
		if err := n.SetAutostart(true); err != nil {
			return errors.Wrapf(err, "setting autostart for network %s", name)
		}
	}

	// always ensure the network is started (active)
	active, err := n.IsActive()
	if err != nil {
		return errors.Wrapf(err, "checking network status for %s", name)
	}

	if !active {
		log.Debugf("network %s is not active, trying to start it...", name)
		if err := n.Create(); err != nil {
			return errors.Wrapf(err, "starting network %s", name)
		}
	}
	return nil
}

func getConnection(connectionURI string) (*libvirt.Connect, error) {
	conn, err := libvirt.NewConnect(connectionURI)
	if err != nil {
		return nil, fmt.Errorf("failed connecting to libvirt socket: %w", lvErr(err))
	}

	return conn, nil
}

// lvErr will return libvirt Error struct containing specific libvirt error code, domain, message and level
func lvErr(err error) libvirt.Error {
	if err != nil {
		if lverr, ok := err.(libvirt.Error); ok {
			return lverr
		}
		return libvirt.Error{Code: libvirt.ERR_INTERNAL_ERROR, Message: "internal error"}
	}
	return libvirt.Error{Code: libvirt.ERR_OK, Message: ""}
}
