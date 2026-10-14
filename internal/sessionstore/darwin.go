//go:build darwin

package sessionstore

import (
	"errors"
	keychain "github.com/keybase/go-keychain"
)

// System stores only the gateway refresh token in the OS credential store.
// macOS uses Security.framework directly, so the secret never enters argv.
type System struct{}

const service = "com.surplustoken.connect.session"
const account = "https://surplustoken.com"

func query() keychain.Item {
	item := keychain.NewItem()
	item.SetSecClass(keychain.SecClassGenericPassword)
	item.SetService(service)
	item.SetAccount(account)
	return item
}
func (System) Load() (string, error) {
	item := query()
	item.SetReturnData(true)
	item.SetMatchLimit(keychain.MatchLimitOne)
	results, err := keychain.QueryItem(item)
	if errors.Is(err, keychain.ErrorItemNotFound) {
		return "", nil
	}
	if err != nil {
		return "", errors.New("无法读取系统钥匙串，请解锁后重试")
	}
	if len(results) == 0 {
		return "", nil
	}
	return string(results[0].Data), nil
}
func (System) Save(token string) error {
	update := keychain.NewItem()
	update.SetData([]byte(token))
	err := keychain.UpdateItem(query(), update)
	if errors.Is(err, keychain.ErrorItemNotFound) {
		item := query()
		item.SetLabel("SurplusToken Connect 登录状态")
		item.SetData([]byte(token))
		err = keychain.AddItem(item)
	}
	if err != nil {
		return errors.New("无法保存登录状态到系统钥匙串")
	}
	return nil
}
func (System) Delete() error {
	err := keychain.DeleteItem(query())
	if err != nil && !errors.Is(err, keychain.ErrorItemNotFound) {
		return errors.New("无法清除系统钥匙串中的登录状态")
	}
	return nil
}
