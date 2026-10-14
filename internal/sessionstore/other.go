//go:build !darwin

package sessionstore

import (
	"errors"
	"github.com/zalando/go-keyring"
)

type System struct{}

const service = "com.surplustoken.connect.session"
const account = "https://surplustoken.com"

func (System) Load() (string, error) {
	token, err := keyring.Get(service, account)
	if errors.Is(err, keyring.ErrNotFound) {
		return "", nil
	}
	if err != nil {
		return "", errors.New("无法读取系统凭据库")
	}
	return token, nil
}
func (System) Save(token string) error {
	if keyring.Set(service, account, token) != nil {
		return errors.New("无法保存登录状态到系统凭据库")
	}
	return nil
}
func (System) Delete() error {
	err := keyring.Delete(service, account)
	if err != nil && !errors.Is(err, keyring.ErrNotFound) {
		return errors.New("无法清除系统凭据库中的登录状态")
	}
	return nil
}
