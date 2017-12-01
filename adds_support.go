package gomixer

import "log"

const (
	defConverter              = "[support] `%s`\n"
	prefixLogConverter        = "[support]"
	ErrorValueNotValidConvert = "не подходящее значение для конвертации"
)

type Support struct {
	Transliter *transliter
	Convert    *convert
	logger     *log.Logger
	StateData  *statelessData
	Former 	   *Form
}

func NewSupport(logger *log.Logger) *Support {
	return &Support{
		Transliter: newTransliter(),
		Convert:    newConverter(logger),
		logger:     logger,
		StateData:  newstatelessData(),

	}
}
