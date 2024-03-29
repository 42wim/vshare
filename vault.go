package main

import (
	"strconv"
	"time"

	"github.com/hashicorp/vault/api"
)

const cubbyPrefix = "vshare"

// vaultInit uses the VAULT_TOKEN to login
func vaultInit() (*api.Client, error) {
	cfg := api.DefaultConfig()

	v, err := api.NewClient(cfg)
	if err != nil {
		return v, err
	}

	return v, nil
}

// setCubby sets the secret in a cubbyhole and returns a response-wrapping token
func setCubby(v *api.Client, secretText, ttl string, encode bool) (string, error) {
	vpath := "cubbyhole/" + cubbyPrefix + strconv.FormatInt(time.Now().UnixNano(), 10)

	_, err := v.Logical().Write(vpath, map[string]interface{}{
		"secret": secretText,
		"encode": encode,
	})
	if err != nil {
		return "", err
	}
	// Set a wrapping lookup function for reads on that path
	v.SetWrappingLookupFunc(func(operation, path string) string {
		return ttl
	})
	// Read again to get a wrapped token
	s, err := v.Logical().Read(vpath)
	if err != nil {
		return "", err
	}

	return s.WrapInfo.Token, nil
}

// unWrap unwraps the received token and returns the secret as a string
func unWrap(token string) (string, bool, error) {
	v, err := vaultInit()
	if err != nil {
		return "", false, err
	}

	s, err := v.Logical().Unwrap(token)
	if err != nil {
		return "", false, err
	}

	return s.Data["secret"].(string), s.Data["encode"].(bool), nil
}
