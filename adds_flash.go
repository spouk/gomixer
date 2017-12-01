package gomixer

import (
	"sync"
	"time"
	"fmt"
	"crypto/md5"
)
const (
	SALT_FLASH_HASH = "#$%dfgdfgWERfvdfgdfgFgSomeSaltHere"
)

type Flash struct {
	sync.RWMutex
	Key   string
	Stock map[string]*FlashMessage
}
type FlashMessage struct {
	Status  string
	Message interface{}
}
//---------------------------------------------------------------------------
//  FLASH:
//---------------------------------------------------------------------------
func newFlash() *Flash{
	n := &Flash{
		Stock: make(map[string]*FlashMessage),
	}
	n.Key = n.generateKey()
	return n
}
func (f *Flash) generateKey() string {
	t := time.Now()
	return fmt.Sprintf("%x", md5.Sum([]byte(t.String()+SALT_FLASH_HASH)))
}
func (f *Flash) Set(status, section string, message interface{}) {
	nm := &FlashMessage{Status: status, Message: message}
	f.Lock()
	f.Stock[section] = nm
	f.Unlock()
}
func (f *Flash) Get(section string) (*FlashMessage) {
	f.Lock()
	defer f.Unlock()
	if result, exists := f.Stock[section]; exists {
		delete(f.Stock, section)
		return result
	}
	return nil
}
func (f *Flash) HaveMsg(section string) bool {
	f.Lock()
	defer f.Unlock()
	_, exists := f.Stock[section]
	if exists {
		return true
	}
	return false
}

