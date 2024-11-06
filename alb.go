package main

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog/log"
	"io"
	"net/http"
	"sync"
)

// AWS ALB load balancer can be used to authorize users via Google.
// If user is authorized, ALB adds a header with JWT token, where
// user email is stored. They JWT is signed by ALB's key. There might
// be several keys, so a key id is also stored in JTW, and we have
// to obtain this key from AWS, and verify the signature.
//
// If all those steps pass, this middleware adds the X-Rmote-User
// header with the user's email.
func AlbUserAuthorizer(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("X-Amzn-Oidc-Data")

		if len(tokenString) == 0 {
			log.Error().Msg("got request with no user claims")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		// Parse the JWT token, and verify it. To verify the cryptographic signature,
		// we need to obtain the signing key, which is specified in the 'kid' field.
		token, err := jwtParser.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Note that the parser has a fixed list of valid signing methods, no need to check it
			if kid, ok := token.Header["kid"]; ok {
				return albKeys.Get(kid.(string))
			}
			return nil, errors.New("no kid field")
		})

		if err != nil {
			log.Error().Err(err).Msg("cannot parse user claims")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if !token.Valid {
			log.Error().Err(err).Msg("user claims token is invalid")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		email := claims["email"].(string)
		log.Info().Msgf("authorized user %s", email)
		r.Header.Set("X-Remote-User", email)
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

var jwtParser = jwt.NewParser(jwt.WithPaddingAllowed(), jwt.WithValidMethods([]string{"ES256"}))

// Obtaining the signing keys from AWS is potentially costly, and we can't do
// that on every request. The definitions below create a cache.
var albKeys = NewAlbKeys()

type AlbKeys struct {
	cache map[string]*ecdsa.PublicKey
	lock  sync.Mutex
}

func NewAlbKeys() *AlbKeys {
	return &AlbKeys{
		cache: make(map[string]*ecdsa.PublicKey),
	}
}

func (self *AlbKeys) Get(kid string) (*ecdsa.PublicKey, error) {
	self.lock.Lock()
	defer self.lock.Unlock()
	if bytes, ok := self.cache[kid]; ok {
		return bytes, nil
	}

	url := fmt.Sprintf("https://public-keys.auth.elb.eu-central-1.amazonaws.com/%v", kid)
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	block, _ := pem.Decode(body)
	if block == nil {
		return nil, errors.New("failed to decode PEM data")
	}
	key, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		return nil, err
	}
	result := key.(*ecdsa.PublicKey)
	self.cache[kid] = result
	return result, nil
}
