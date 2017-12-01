package gomixer

type GomixerMiddleware func(handler HandlerFunc) HandlerFunc
type Gomiddlestock map[string][]GomixerMiddleware

//func(m *Gomiddlestock) Add()
