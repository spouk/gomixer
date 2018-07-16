package gomixer

import (
	"os"
	"log"
	"encoding/gob"
	"io"
	"io/ioutil"
	"github.com/go-yaml/yaml"
)

const (
    logConfigPrefix           = "[configurator] "
    logConfigFlags            = log.Ldate | log.Ltime | log.Lshortfile
    successReadConfigFile     = "файл конфигурации успешно прочитан"
    msgsuccessReadDumpConfig  = "дамп конфига успешно прочитан"
    msgsuccessWriteDumpConfig = "дамп конфига успешно записан"
    msgerrorWriteDumpConfig   = "дамп конфига записан с ошибкой"

    dumpfileflag = os.O_CREATE | os.O_RDWR | os.O_TRUNC
    dumpfileperm = 0666
)

type configurator struct {
    logger             *log.Logger
    pathDumpConfigFile string
    Config             interface{}
}

func NewConfigurator(fileConfigYaml string, logout io.Writer, config interface{}) *configurator {
    //создаю инстанс конфигуратора
    c := &configurator{
	Config:config,
    }
    //создаю логгер
	if logout == nil {
		c.logger = log.New(os.Stdout, logConfigPrefix, logConfigFlags)
	} else {
		c.logger = log.New(logout, logConfigPrefix, logConfigFlags)
	}
	//открываю файл конфигурации
	f, err := os.Open(fileConfigYaml)
	if err != nil {
		panic(err)
	}
	//читаю файл конфига
	b, err := ioutil.ReadAll(f)
	if err != nil {
		panic(err)
	}
	//конвертирую его в структуру
	err = yaml.Unmarshal(b, c.Config)
	if err != nil {
		panic(err)
	}
	c.logger.Printf(successReadConfigFile)

	//возвращаю результат
	return c
}

//запись дампа конфигурационного файла
func (c *configurator) WriteDumpConfig() error {
	f, err := os.OpenFile(c.pathDumpConfigFile, dumpfileflag, dumpfileperm)
	if err != nil {
		return err
	}
	defer f.Close()
	enc := gob.NewEncoder(f)
	err = enc.Encode(c.Config)
	if err != nil {
		return err
	}
	if c.logger != nil {
		c.logger.Printf(msgsuccessWriteDumpConfig)
	}
	return nil
}

//читаем дамп конфигурационного файла
func (c *configurator) ReadDumpConfig() (error) {
	f, err := os.Open(c.pathDumpConfigFile)
	if err != nil {
		return err
	}
	defer f.Close()
	dec := gob.NewDecoder(f)
	err = dec.Decode(c.Config)
	if err != nil {
		return err
	}
	c.logger.Printf(msgsuccessReadDumpConfig)
	return nil
}
