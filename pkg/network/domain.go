package network

import (
	"encoding/xml"
	"fmt"

	"github.com/pkg/errors"
	"libvirt.org/go/libvirt"

	"github.com/day0ops/netctl/pkg/log"
)

func (n *Network) checkDomains(conn *libvirt.Connect) error {
	type source struct {
		// XMLName xml.Name `xml:"source"`
		Network string `xml:"network,attr"`
	}
	type iface struct {
		// XMLName xml.Name `xml:"interface"`
		Source source `xml:"source"`
	}
	type result struct {
		// XMLName xml.Name `xml:"domain"`
		Name       string  `xml:"name"`
		Interfaces []iface `xml:"devices>interface"`
	}

	// iterate over every (also turned off) domains, and check if it
	// is using the private network. Do *not* delete the network if
	// that is the case
	log.Debug("trying to list all domains...")
	doms, err := conn.ListAllDomains(0)
	if err != nil {
		return errors.Wrap(err, "list all domains")
	}
	log.Debugf("listed all domains: total of %n domains", len(doms))

	// fail if there are 0 domains
	if len(doms) == 0 {
		log.Warn("list of domains is 0 length")
	}

	for _, dom := range doms {
		// get the name of the domain we iterate over
		log.Debug("trying to get name of domain...")
		name, err := dom.GetName()
		if err != nil {
			return errors.Wrap(err, "failed to get name of a domain")
		}
		log.Debugf("got domain name: %s", name)

		// unfortunately, there is no better way to retrieve a list of all defined interfaces
		// in domains than getting it from the defined XML of all domains
		// NOTE: conn.ListAllInterfaces does not help in this case
		log.Debugf("getting XML for domain %s...", name)
		xmlString, err := dom.GetXMLDesc(libvirt.DOMAIN_XML_INACTIVE)
		if err != nil {
			return errors.Wrapf(err, "failed to get XML of domain '%s'", name)
		}
		log.Debugf("got XML for domain %s", name)

		v := result{}
		err = xml.Unmarshal([]byte(xmlString), &v)
		if err != nil {
			return errors.Wrapf(err, "failed to unmarshal XML of domain '%s", name)
		}
		log.Debugf("unmarshaled XML for domain %s: %#v", name, v)

		// iterate over the found interfaces
		for _, i := range v.Interfaces {
			if i.Source.Network == n.Name {
				log.Debugf("domain %s DOES use network %s, aborting...", name, n.Name)
				return fmt.Errorf("network still in use at least by domain '%s'", name)
			}
			log.Debugf("domain %s does not use network %s", name, n.Name)
		}
	}

	return nil
}
