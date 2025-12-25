package middlewares

import (
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type JWKS struct {
	Keys []JSONWebKey `json:"keys"`
}

type JSONWebKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

var (
	keycloakPublicKeys *JWKS
	lastRefresh        time.Time
	mu                 sync.Mutex
)

func fetchKeycloakJWKS() (*JWKS, error) {
	keycloakURL := os.Getenv("KEYCLOAK_ISSUER_URL")
	if !strings.HasSuffix(keycloakURL, "/") {
		keycloakURL += "/"
	}
	jwksURL := keycloakURL + "protocol/openid-connect/certs"

	resp, err := http.Get(jwksURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch JWKS: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("non-200 response: %d", resp.StatusCode)
	}

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return nil, fmt.Errorf("failed to decode JWKS: %v", err)
	}
	return &jwks, nil
}

func getKeyFromKid(kid string) (*JSONWebKey, error) {
	mu.Lock()
	defer mu.Unlock()

	if keycloakPublicKeys == nil || time.Since(lastRefresh) > 5*time.Minute {
		newJWKS, err := fetchKeycloakJWKS()
		if err != nil {
			return nil, err
		}
		keycloakPublicKeys = newJWKS
		lastRefresh = time.Now()
	}

	for _, key := range keycloakPublicKeys.Keys {
		if key.Kid == kid {
			return &key, nil
		}
	}
	return nil, fmt.Errorf("kid not found: %s", kid)
}

func jsonKeyToPublicKey(jwk *JSONWebKey) (*rsa.PublicKey, error) {
	if jwk.Kty != "RSA" {
		return nil, fmt.Errorf("unsupported key type: %s", jwk.Kty)
	}

	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode n: %v", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode e: %v", err)
	}

	var eInt int
	if len(eBytes) < 4 {
		tmp := 0
		for _, b := range eBytes {
			tmp = tmp<<8 | int(b)
		}
		eInt = tmp
	} else {
		eInt = int(binary.BigEndian.Uint32(eBytes))
	}

	pub := &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: eInt,
	}

	if pub.E < 2 {
		return nil, fmt.Errorf("exponent %d is too small; must be >= 2", pub.E)
	}

	if pub.N.Cmp(big.NewInt(1)) <= 0 {
		return nil, fmt.Errorf("invalid modulus; N must be > 1")
	}

	if pub.N.BitLen() == 0 {
		return nil, fmt.Errorf("invalid modulus; bit length is 0")
	}

	return pub, nil
}
