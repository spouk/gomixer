//---------------------------------------------------------------------------
//  конвертер - пакет для помоши в конвертации разных величин и приведения их
//	к приемлемому типу для ситуации
//---------------------------------------------------------------------------

package gomixer

import (
	"log"
	"time"
	"strings"
	"reflect"
	"math/rand"
	"strconv"

)

type convert struct {
	logger  *log.Logger
	value   interface{}
	result  interface{}
	stockFu map[string]func()
}

var (
	acceptTypes []interface{} = []interface{}{
		"", 0, int64(0),
	}
)

func newConverter(log *log.Logger) *convert {
	f := &convert{
		stockFu: make(map[string]func()),
		logger:  log,
	}
	f.stockFu["string"] = f.stringToInt
	f.stockFu["string"] = f.stringToInt64
	return f
}
func (m *convert) StrToInt() (*convert) {
	if f, exists := m.stockFu["string"]; exists {
		f()
	}
	return m
}
func (m *convert) StrToInt64() (*convert) {
	if f, exists := m.stockFu["string"]; exists {
		f()
	}
	return m
}

//---------------------------------------------------------------------------
//  String to Int64
//---------------------------------------------------------------------------
func (m *convert) stringToInt64() {
	m.stringToInt()
	if m.result != nil {
		m.result = int64(m.result.(int))
	} else {
		m.result = nil
	}
}

//---------------------------------------------------------------------------
//  String to Int
//---------------------------------------------------------------------------
func (m *convert) stringToInt() {
	if r, err := strconv.Atoi(m.value.(string)); err != nil {
		m.logger.Printf(defConverter, err.Error())
		m.result = nil
	} else {
		m.result = r
	}
}

//---------------------------------------------------------------------------
//  возвращает результат конвертации
//---------------------------------------------------------------------------
func (m *convert) Result() interface{} {
	return m.result
}

//---------------------------------------------------------------------------
//  инциализация вводным значением
//---------------------------------------------------------------------------
func (m *convert) Value(value interface{}) (*convert) {
	if m.checkValue(value) {
		m.value = value
		return m
	}
	return nil
}

//---------------------------------------------------------------------------
//  проверка типа поступившего значения на возможность конвертации
//---------------------------------------------------------------------------
func (m *convert) checkValue(value interface{}) bool {
	tValue := reflect.TypeOf(value)
	for _, x := range acceptTypes {
		if tValue == reflect.TypeOf(x) {
			return true
		}
	}
	m.logger.Printf(defConverter, ErrorValueNotValidConvert)
	return false
}

func (m *convert) FloatToString(input_num float64) string {
	// to convert a float number to a string
	return strconv.FormatFloat(input_num, 'f', 2, 64)

}
func (m *convert) Int64ToString(input_num int64) string {
	return strconv.FormatInt(input_num, 10)

}
func (m *convert) DirectStringtoInt64(v string) int64 {
	if res, err := strconv.Atoi(v); err != nil {
		m.logger.Printf(defConverter, err.Error())
		return 0
	} else {
		return int64(res)
	}
}
func (m *convert) DirectStringtoIntSlice(a []string) []int {
	var result []int
	if len(a) > 0 {
		for _, x := range a {
			if res, err := strconv.Atoi(x); err != nil {
				m.logger.Printf(defConverter, err.Error())
				continue
			} else {
				result = append(result, res)
			}
		}
	}
	return result
}
func (m *convert) DirectStringtoInt64Slice(a []string) []int64 {
	var result []int64
	if len(a) > 0 {
		for _, x:= range a {
			if res, err := strconv.Atoi(x); err != nil {
				m.logger.Printf(defConverter, err.Error())
				continue
			} else {
				result = append(result, int64(res))
			}
		}
	}
	return result
}
func (m *convert) DirectStringFormtoBool(v string) bool{
	if v == "" {
		return false
	}
	return true
}
func (m *convert) DirectStringtoInt(v string) int {
	if len(v) > 0 {
		if res, err := strconv.Atoi(v); err != nil {
			m.logger.Printf(defConverter, err.Error())
			return 0
		} else {
			return res
		}
	}
	return 0

}
func (m *convert) DirectStringtoFloat64(v string) float64 {
	if res, err := strconv.ParseFloat(v, 10); err != nil {
		m.logger.Printf(defConverter, err.Error())
		return 0
	} else {
		return res
	}
}

//---------------------------------------------------------------------------
//  time convert
//---------------------------------------------------------------------------
func (m *convert) ConvertHTMLDatetoUnix(date string) (int64, error) {
	if len(date) > 0 {
		result, err := time.Parse("2006-01-02", date)
		if err == nil {
			return result.Unix(), err
		} else {
			return 0, err
		}
	}
	return 0, nil

}
func (m *convert) ConvertUnixTimeToString(unixtime int64) string {
	return time.Unix(unixtime, 0).String()
}

//convert timeUnix->HTML5Datatime_local(string)
func (m *convert) TimeUnixToDataLocal(unixtime int64) string {
	tmp_result := time.Unix(unixtime, 0).Format(time.RFC3339)
	g := strings.Join(strings.SplitAfterN(tmp_result, ":", 3)[:2], "")
	return g[:len(g)-1]
}

//convert HTML5Datatime_local(string)->TimeUnix
func (m *convert) DataLocalToTimeUnix(datatimeLocal string) int64 {
	r, _ := time.Parse(time.RFC3339, datatimeLocal+":00Z")
	return r.Unix()
}

//convert HTML5Data->UnixTime
func (m *convert) HTML5DataToUnix(s string) int64 {
	l := "2006-01-02"
	r, _ := time.Parse(l, s)
	return r.Unix()
}

//UnixTime->HTML5Data
func (m *convert) UnixtimetoHTML5Date(unixtime int64) string {
	return time.Unix(unixtime, 0).Format("2006-01-02")
}

//рандомный генератор строк переменной длины
func (m *convert) RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
