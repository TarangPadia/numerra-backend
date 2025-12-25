package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"sort"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func VerifyShopifyHMACFromFiberQuery(c *fiber.Ctx, secret string) bool {
	qa := c.Context().QueryArgs()

	params := make([][2]string, 0, qa.Len())
	qa.VisitAll(func(k, v []byte) {
		key := string(k)
		if key == "hmac" || key == "signature" {
			return
		}
		params = append(params, [2]string{key, string(v)})
	})

	sort.Slice(params, func(i, j int) bool { return params[i][0] < params[j][0] })

	var parts []string
	for _, kv := range params {
		parts = append(parts, kv[0]+"="+url.QueryEscape(kv[1]))
	}
	msg := strings.Join(parts, "&")

	gotHmac := c.Query("hmac")
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(msg))
	expected := hex.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(gotHmac))
}
