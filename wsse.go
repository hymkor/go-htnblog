package htnblog

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
)

func newNonce() (string, error) {
	uuid1, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	h := sha1.New()
	io.WriteString(h, uuid1.String())
	return base64.StdEncoding.EncodeToString(h.Sum(nil)), nil
	//bin1, err := uuid1.MarshalBinary()
	//if err != nil {
	//	return "", err
	//}
	//return base64.StdEncoding.EncodeToString(bin1), nil
}

func newWsse(req *http.Request, username, apiKey, nonce string) (string, error) {
	created := time.Now().UTC().Format("2006-01-02T15:04:05Z")

	nonce, err := newNonce()
	if err != nil {
		return "", err
	}

	h := sha1.New()
	io.WriteString(h, nonce+created+apiKey)
	digest := base64.StdEncoding.EncodeToString(h.Sum(nil))

	// X-WSSE: UsernameToken Username="hatena", PasswordDigest="ZCNaK2jrXr4+zsCaYK/YLUxImZU=", Nonce="Uh95NQlviNpJQR1MmML+zq6pFxE=", Created="2005-01-18T03:20:15Z"

	// wtm4akm9xe.wbij775cbb77g.draft@blog.hatena.ne.jp

	return fmt.Sprintf(
		`UsernameToken `+
			`Username="%s", `+
			`PasswordDigest="%s", `+
			`Nonce="%s", `+
			`Created="%s"`,
		username,
		digest,
		base64.StdEncoding.EncodeToString([]byte(nonce)),
		created), nil
}
