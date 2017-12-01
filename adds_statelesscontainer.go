//---------------------------------------------------------------------------
//  контейнер для сохранения состояние чего-нибудь между запросами
//---------------------------------------------------------------------------

package gomixer

import (
	"sync"

)

type statelessData struct {
	sync.RWMutex
	Data map[string]*Box
}
type Box struct {
	Key  string
	Data map[string]interface{}
}

func newstatelessData() *statelessData {
	return &statelessData{
		Data: make(map[string]*Box),
	}
}
func (s *statelessData) NewBox(key string) *Box {
	return &Box{
		Key:  key,
		Data: make(map[string]interface{}),
	}
}
func (s *statelessData) Save(key string, value *Box) {
	s.Lock()
	defer s.Unlock()
	s.Data[key] = value
}
func (s *statelessData) Get(key string) *map[string]interface{} {
	s.Lock()
	defer s.Unlock()
	value, exists := s.Data[key]
	if !exists {
		return nil
	}
	return &value.Data
}
func (s *Box) Reset() {
	s.Data = make(map[string]interface{})
	return
}
