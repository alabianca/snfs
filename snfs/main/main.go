package main

import (
	"flag"
	"fmt"
	"strings"

	"github.com/alabianca/snfs/snfs/discovery"

	"github.com/alabianca/snfs/snfs/server"
)

var port = flag.Int("p", 4000, "Port of the server")
var instance = flag.String("i", "default", "Instance Name")

const topLevelDomain = ".snfs.com"

func main() {
	flag.Parse()
	addr := ""

	server := server.Server{
		Port: *port,
		Addr: addr,
	}

	server.SetDiscoveryManager(
		discovery.MdnsStrategy(mdnsConfig),
	)

	server.StartClientConnectivityService()

	if err := server.HTTPListenAndServe(); err != nil {
		panic(err)
	}
}

func mdnsConfig(m *discovery.MdnsService) {
	instanceName, _ := splitFromTopLevelDomain(*instance)
	m.SetPort(*port)
	m.SetInstance(instanceName)

}

func splitFromTopLevelDomain(instance string) (string, error) {
	if !strings.Contains(instance, topLevelDomain) {
		return "", fmt.Errorf("Instance does not contain top level domain")
	}
	split := strings.Split(instance, topLevelDomain)

	return split[0], nil
}
