package utils

import (
	"errors"
	"net"
	"os"
	"regexp"
	"strings"
)

var ipRegexp = regexp.MustCompile(`\d+\.\d+\.\d+\.\d+/\d+`)

func LocalIpByName(name string) (string, error) {
	itf, err := net.InterfaceByName(name)
	if err != nil {
		return "", err
	}
	addrs, err := itf.Addrs()
	if err != nil {
		return "", err
	}
	if len(addrs) == 0 {
		return "", errors.New("Got addr Error ")
	}
	for _, addr := range addrs {
		if ipRegexp.MatchString(addr.String()) {
			return strings.Split(addr.String(), "/")[0], nil
		}
	}
	return "", errors.New("Got addr Error ")

}

func GetHost() (string, error) {
	name, err := os.Hostname()
	return name, err
}
