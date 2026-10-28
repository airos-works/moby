//go:build haiku

package x509

import "errors"

func (c *Certificate) systemVerify(opts *VerifyOptions) (chains [][]*Certificate, err error) {
	return nil, errors.New("x509: system verification not supported on haiku")
}

func loadSystemRoots() (*CertPool, error) {
	// Haiku keeps CA certificates in /boot/system/data/ssl; the caller can
	// supply roots through VerifyOptions.Roots until this is wired up.
	return nil, nil
}
