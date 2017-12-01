package gomixer

import (
	"time"
	"fmt"
	"crypto/sha1"
	"crypto/md5"
	"math/rand"
)

const (
	CSRF_LENGTH_KEY  = 7
	CSRF_SALT        = "Cvdfg345DFg234@#$dfgxcvbq"
	CSRF_ACTIVE_TIME = 1
)

type CSRF struct {
	TimeActive time.Duration
	TimeStart  time.Time
	Salt       string
	Key        string
	ReadyKey   string
	Csrf_form  func() *string
	Csrf_head  func() *string
}
func NewCSRF(minutesActive int, salt string) (*CSRF) {
	n := &CSRF{
		TimeActive: time.Duration(minutesActive) * time.Minute,
		TimeStart:  time.Now(),
		Salt:       salt,
	}
	if salt == "" {
		n.Salt =  CSRF_SALT
	}
	n.Key = n.randomGenerate(CSRF_LENGTH_KEY)
	//n.ReadyKey  = fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%v%v", n.Key, n.Salt))))
	n.Csrf_form = n.wrapper(true, false)
	n.Csrf_head = n.wrapper(false, true)
	return n
}

func (c *CSRF) wrapper(form, head bool) (func() (*string)) {
	return func() (*string) {
		//получение точного времени истекания действия токена
		_tmptime := c.TimeStart.Add(c.TimeActive)
		//провера на истекший период, если истек, создаю новый ключ, если нет использую старый
		if _tmptime.Before(time.Now()) {
			c.Key = c.randomGenerate(CSRF_LENGTH_KEY)
			c.ReadyKey  = fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%v%v", c.Key, c.Salt))))
			c.TimeStart = time.Now()
			fmt.Printf("[csrf] `создаю новый ключ и обновляю время %v:%v\n", c.Key, c.ReadyKey)
		} else {
			if c.ReadyKey == "" {
				c.ReadyKey  = fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%v%v", c.Key, c.Salt))))
			} else {
				fmt.Printf("[csrf] `использую старый ключ `%v`:%v\n",c.Key,  c.ReadyKey)
			}
		}
		var result string
		if form {
			result = fmt.Sprintf(`<input type="hidden" name="csrf_token" value="%s"> `, c.ReadyKey)
		} else if head {
			result = fmt.Sprintf(`<meta id="csrf_token_ajax" content="%s" name="csrf_token_ajax" />`, c.ReadyKey)
		}
		return &result
	}
}
func (c CSRF) VerifyToken(s *Carry) bool {
	var token string
	if s.r.Method == "GET" {
		token = s.FormGetValue("csrf_token")
	}
	if s.r.Method == "POST" {
		token = s.FormPostGetValue("csrf_token")
	}
	if token == c.ReadyKey {
		return true
	}
	return false
}
func (c CSRF) VerifyTokenString(token string) bool {
	if token == c.ReadyKey {
		return true
	}
	return false
}
func (c CSRF) randomGenerate(count int) string {
	randInt := func(min, max int) int {
		return min + rand.Intn(max-min)
	}
	rand.Seed(time.Now().UTC().UnixNano())
	bytes := make([]byte, count)
	for i := 0; i < count; i++ {
		bytes[i] = byte(randInt(30, 90))
	}
	h := sha1.New()
	h.Write(bytes)
	return fmt.Sprintf("%x", h.Sum(nil))
}
