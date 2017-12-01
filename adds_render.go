package gomixer

import (
	"html/template"
	"io"
	"fmt"
	"strings"
	"net/http"
	"sync"
	"bytes"
	"time"
	"math/rand"
	"os"
	"io/ioutil"
	"reflect"
	"encoding/json"
	"log"
)

const (
	ERROR_HTTPMETHODNOTACCEPT = "http method not allowed "
	ERROR_READTEMPLATES       = "%s"
	ERROR_READ_TXTFILE        = "%s"
)

var (
	Local_filter = map[string]interface{}{
		"random":  RandomGenerator,
		"count":   strings.Count,
		"split":   strings.Split,
		"title":   strings.Title,
		"lower":   strings.ToLower,
		"totitle": strings.ToTitle,
		"makemap": MakeMap,
		"in":      MapIn,
		"andlist": AndList,
		"upper":   strings.ToUpper,
		"concat": Concat,

		"unixtime":       UnixtimeNormal,
		"unixtimeformat": UnixtimeNormalFormatData,
		"unixtodata":     UnixtimeNormalFormatData,

		"yesno": YesNo,
		"html2": func(value string) template.HTML {
			return template.HTML(fmt.Sprint(value))
		},
		"type":        TypeIs,
		"jsonconvert": JSONconvert,
	}
)
//---------------------------------------------------------------------------
//  MIXERRENDERDEFAULT: определение типа рендера
//---------------------------------------------------------------------------
type (
	Render struct {
		sync.RWMutex
		Temp    *template.Template
		Filters template.FuncMap
		Debug   bool
		Path    string
		logger  *log.Logger
	}
)

func NewRender(path string, debug bool, logger *log.Logger) *Render {
	sf := &Render{}
	defer sf.catcherPanic()
	sf.Filters = template.FuncMap{}
	sf.AddFilters(Local_filter)
	sf.Path = path
	sf.Debug = debug
	if logger != nil {
		sf.logger = logger
	} else {
		sf.logger = log.New(os.Stdout, PREFIXLOGGER, log.Ltime|log.Ldate|log.Lshortfile)
	}
	return sf
}
func (s *Render) AddUserFilter(name string, f interface{}) {
	s.Filters[name] = f
}
func (s *Render) AddFilters(stack map[string]interface{}) {
	for k, v := range stack {
		s.Filters[k] = v
	}
}
func (s *Render) ReloadTemplate() {
	defer s.catcherPanic()
	if s.Debug || s.Temp == nil {
		s.Temp = template.Must(template.New("indexstock").Funcs(s.Filters).ParseGlob(s.Path))
	}
}
func (s *Render) Render(name string, data interface{}, w io.Writer) (err error) {
	defer s.catcherPanic()
	if s.Debug || s.Temp == nil {
		s.ReloadTemplate()
	}
	buf := new(bytes.Buffer)
	if err = s.Temp.ExecuteTemplate(buf, name, data); err != nil {
		s.logger.Printf(fmt.Sprintf(ERROR_READTEMPLATES, err.Error()))
		return
	}
	resp := w.(http.ResponseWriter)
	resp.Header().Add(ContentType, TextHTMLCharsetUTF8)
	resp.WriteHeader(http.StatusOK)
	resp.Write(s.HTMLTrims(buf.Bytes()))

	return
}
func (s *Render) RenderCode(httpCode int, name string, data interface{}, w io.Writer) (err error) {
	defer s.catcherPanic()
	if s.Debug || s.Temp == nil {
		s.ReloadTemplate()
	}
	buf := new(bytes.Buffer)
	if err = s.Temp.ExecuteTemplate(buf, name, data); err != nil {
		s.logger.Printf(fmt.Sprintf(ERROR_READTEMPLATES, err.Error()))
		return
	}
	resp := w.(http.ResponseWriter)
	resp.Header().Add(ContentType, TextHTMLCharsetUTF8)
	resp.WriteHeader(httpCode)
	resp.Write(s.HTMLTrims(buf.Bytes()))

	return
}
func (s *Render) RenderTxt(httpCode int, name string, w io.Writer) (err error) {
	//read txt file
	file, err := os.Open(name)
	if err != nil {
		s.logger.Printf(ERROR_READ_TXTFILE, err.Error())
		return err
	}
	outFile, err := ioutil.ReadAll(file)
	if err != nil {
		s.logger.Printf(ERROR_READ_TXTFILE, err.Error())
		return err
	}
	resp := w.(http.ResponseWriter)
	resp.Header().Add(ContentType, TextPlain)
	resp.WriteHeader(httpCode)
	resp.Write(outFile)

	return
}
func (s *Render) catcherPanic() {
	msgPanic := recover()
	if msgPanic != nil && s.logger != nil {
		s.logger.Printf("[ERROR TEMPLATE] %v", msgPanic)
	}
}
func (s *Render) HTMLTrims(body []byte) []byte {
	result := []string{}
	for _, line := range strings.Split(string(body), "\n") {
		if len(line) != 0 && len(strings.TrimSpace(line)) != 0 {
			result = append(result, line)
		}
	}
	return []byte(strings.Join(result, "\n"))
}
func (s *Render) ShowStack() {
	fmt.Println(s.Filters)
}

//возращает тип аргумента
func TypeIs(value interface{}) string {
	v := reflect.ValueOf(value)
	var result string
	switch v.Kind() {
	case reflect.Bool:
		result = "bool"
	case reflect.Int, reflect.Int8, reflect.Int32, reflect.Int64:
		result = "integer"
	case reflect.Uint, reflect.Uint8, reflect.Uint32, reflect.Uint64:
		result = "unsigned integer"
	case reflect.Float32, reflect.Float64:
		result = "float"
	case reflect.String:
		result = "string"
	case reflect.Slice:
		result = "slice"
	case reflect.Map:
		result = "map"
	case reflect.Chan:
		result = "chan"
	default:
		result = "undefine type"
	}
	return result
}
func MapIn(value interface{}, stock interface{}) bool {
	switch value.(type) {
	case int64:
		for _, x := range stock.([]int64) {
			if x == value.(int64) {
				return true
			}
		}
	case int:
		for _, x := range stock.([]int) {
			if x == value.(int) {
				return true
			}
		}
	case string:
		for _, x := range stock.([]string) {
			if x == value.(string) {
				return true
			}
		}

	}
	return false
}
func MakeMap(value ...string) ([]string) {
	return value
}
func AndList(listValues ...interface{}) (bool) {
	for _, v := range listValues {
		if v == nil {
			return false
		}
	}
	return true
}
func YesNo(value bool, yes, no string) string {
	if value {
		return yes
	}
	return no
}

//---------------------------------------------------------------------------
//  TIME Functions
//---------------------------------------------------------------------------
func UnixtimeNormal(unixtime int64) string {
	return time.Unix(unixtime, 0).String()
}

//UnixTime->HTML5Data
func UnixtimeNormalFormatData(unixtime int64) string {
	return time.Unix(unixtime, 0).Format("2006-01-02")
}

//convert HTML5Data->UnixTime
func HTML5DataToUnix(s string) int64 {
	l := "2006-01-02"
	r, _ := time.Parse(l, s)
	return r.Unix()
}

//convert timeUnix->HTML5Datatime_local(string)
func TimeUnixToDataLocal(unixtime int64) string {
	tmp_result := time.Unix(unixtime, 0).Format(time.RFC3339)
	g := strings.Join(strings.SplitAfterN(tmp_result, ":", 3)[:2], "")
	return g[:len(g)-1]
}

//convert HTML5Datatime_local(string)->TimeUnix
func DataLocalToTimeUnix(datatimeLocal string) int64 {
	r, _ := time.Parse(time.RFC3339, datatimeLocal+":00Z")
	return r.Unix()
}

//---------------------------------------------------------------------------
//  RANDOM for Update Css and JS file in head pages
//---------------------------------------------------------------------------
func RandomGenerator() int {
	return rand.Intn(1000)
}

//---------------------------------------------------------------------------
//  JSON CONVERT SUPPORT TEMPALTES
//---------------------------------------------------------------------------
func JSONconvert(obj interface{}) string {
	buf, err := json.Marshal(obj)
	if err != nil {
		fmt.Printf(err.Error())
		return ""
	}
	return string(buf)
}
func Concat(s, s1 string ) string {
	return s + s1
}
