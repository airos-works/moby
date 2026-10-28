//go:build haiku

package listeners

import (
	"crypto/tls"
	"net"
	"os"

	"github.com/docker/go-connections/sockets"
	"github.com/pkg/errors"
)

// Init creates new listeners for the daemon, without systemd socket
// activation support ("fd" listeners error on Haiku).
func Init(proto, addr, socketGroup string, tlsConfig *tls.Config) ([]net.Listener, error) {
	ls := []net.Listener{}

	switch proto {
	case "fd":
		return nil, errors.New("socket-activated listeners are not supported on haiku")
	case "tcp":
		l, err := sockets.NewTCPSocket(addr, tlsConfig)
		if err != nil {
			return nil, err
		}
		ls = append(ls, l)
	case "unix":
		gid, err := lookupGID(socketGroup)
		if err != nil {
			gid = os.Getgid()
		}
		l, err := sockets.NewUnixSocket(addr, gid)
		if err != nil {
			return nil, errors.Wrapf(err, "can't create unix socket %s", addr)
		}
		ls = append(ls, l)
	default:
		return nil, errors.Errorf("invalid protocol format: %q", proto)
	}

	return ls, nil
}

