package common

import (
	"github.com/chmike/securecookie"
)

var CookieSettings = map[string]securecookie.Params{}

func setupCookies(list map[string]securecookie.Params) error {
	App.SecureCookie = make(map[string]*securecookie.Obj)
	App.SecureCookieKey = securecookie.MustGenerateRandomKey()
	CookieSettings = list
	for k, v := range CookieSettings {
		sc, err := securecookie.New(k, App.SecureCookieKey, v)
		if err != nil {
			return err
		}
		App.SecureCookie[k] = sc
	}
	return nil
}

func ReplenishExpiredCookie(key string, t int) *securecookie.Obj {
	p := CookieSettings[key]
	p.MaxAge = t
	return securecookie.MustNew(key, App.SecureCookieKey, p)
}
