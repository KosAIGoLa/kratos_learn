package captcha

import (
	"time"

	"github.com/mojocn/base64Captcha"
)

var store = base64Captcha.NewMemoryStore(base64Captcha.GCLimitNumber, 10*time.Minute)

type CaptchaManager struct {
	store base64Captcha.Store
}

func NewCaptchaManager() *CaptchaManager {
	return &CaptchaManager{
		store: store,
	}
}

func (c *CaptchaManager) Generate() (id, b64s string, err error) {
	driver := base64Captcha.NewDriverDigit(80, 240, 5, 0.7, 80)
	captcha := base64Captcha.NewCaptcha(driver, c.store)
	id, b64s, _, err = captcha.Generate()
	return id, b64s, err
}

func (c *CaptchaManager) Verify(id, answer string, clear bool) bool {
	return c.store.Verify(id, answer, clear)
}
