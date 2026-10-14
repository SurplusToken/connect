package configuration

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/tidwall/jsonc"
	"io"
)

// Duplicate keys are ambiguous across clients/parsers. Refuse them rather than
// editing the first occurrence while the client reads the last one.
func validateJSON(raw []byte) error {
	raw = bytes.TrimPrefix(raw, []byte("\xef\xbb\xbf"))
	decoder := json.NewDecoder(bytes.NewReader(jsonc.ToJSON(raw)))
	decoder.UseNumber()
	var consume func() error
	consume = func() error {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		switch token {
		case json.Delim('{'):
			seen := map[string]bool{}
			for decoder.More() {
				key, err := decoder.Token()
				if err != nil {
					return err
				}
				name, ok := key.(string)
				if !ok || seen[name] {
					return errors.New("重复或无效 JSON 字段")
				}
				seen[name] = true
				if err = consume(); err != nil {
					return err
				}
			}
			token, err = decoder.Token()
			if err != nil || token != json.Delim('}') {
				return errors.New("无效 JSON 对象")
			}
		case json.Delim('['):
			for decoder.More() {
				if err := consume(); err != nil {
					return err
				}
			}
			token, err = decoder.Token()
			if err != nil || token != json.Delim(']') {
				return errors.New("无效 JSON 数组")
			}
		}
		return nil
	}
	if err := consume(); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return errors.New("多个 JSON 文档")
	}
	return nil
}
