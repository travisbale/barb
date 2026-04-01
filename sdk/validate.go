package sdk

import (
	"fmt"
	"net"
)

func (r EnrollMiragedRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name: required")
	}
	if r.Address == "" {
		return fmt.Errorf("address: required")
	}
	if _, _, err := net.SplitHostPort(r.Address); err != nil {
		return fmt.Errorf("address: must be in host:port format")
	}
	if r.SecretHostname == "" {
		return fmt.Errorf("secret_hostname: required")
	}
	if r.Token == "" {
		return fmt.Errorf("token: required")
	}
	return nil
}
