package surl

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// PathFormatter includes the signature and expiry in a
// message according to the format: <sig>.<exp>/<data>. Suitable for
// URL paths as an alternative to using query parameters.
type PathFormatter struct {
	signer *Signer
}

// AddExpiry adds expiry as a path component e.g. /foo/bar ->
// 390830893/foo/bar
func (f *PathFormatter) AddExpiry(unsigned *url.URL, expiry time.Time) {
	// convert expiry to string
	expBytes := strconv.FormatInt(expiry.Unix(), 10)

	unsigned.Path = expBytes + unsigned.Path
}

// AddSignature adds signature as a path component alongside the expiry e.g.
// abZ3G/foo/bar -> /KKLJjd3090fklaJKLJK.abZ3G/foo/bar
func (f *PathFormatter) AddSignature(payload *url.URL, sig []byte) {
	encoded := base64.RawURLEncoding.EncodeToString(sig)

	payload.Path = "/" + encoded + "." + payload.Path
}

// ExtractSignature decodes and splits the signature and payload from the signed message.
func (f *PathFormatter) ExtractSignature(u *url.URL) (*url.URL, []byte, error) {
	// prise apart encoded and payload
	encoded, payload, found := strings.Cut(u.Path, ".")
	if !found {
		return nil, nil, ErrInvalidSignedURL
	}
	// remove leading /
	encoded = encoded[1:]

	sig, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: invalid base64: %s", ErrInvalidSignature, encoded)
	}

	u.Path = payload

	if f.signer.skipQuery {
		// remove all query params because they don't form part of the input to
		// the signature computation
		u.RawQuery = ""
	}

	return u, sig, nil
}

// ExtractExpiry decodes and splits the expiry and data from the payload.
func (*PathFormatter) ExtractExpiry(u *url.URL) (*url.URL, time.Time, error) {
	// prise apart expiry and data
	expiry, path, found := strings.Cut(u.Path, "/")
	if !found {
		return nil, time.Time{}, ErrInvalidSignedURL
	}
	// add leading slash back to path
	u.Path = "/" + path

	// convert bytes into int
	expInt, err := strconv.ParseInt(string(expiry), 10, 64)
	if err != nil {
		return nil, time.Time{}, err
	}
	// convert int into time.Time
	t := time.Unix(expInt, 0)

	return u, t, nil
}