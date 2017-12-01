//---------------------------------------------------------------------------
// обеспечивает создание передачи контекста сквозь череду миддлов и запроса
// в контексте может быть что угодно, контейнеры и другая информация требуемая
// для корректной обработки запроса
// в довесок дает доступ к открытому функционалу мультиплексора и расширяющих
// функциональность структур, по типу flash, csrf и прочее
//---------------------------------------------------------------------------

package gomixer

import (
	"net/http"
	"mime/multipart"
	"log"
	"context"
	"errors"
	"io"
	"os"
	"encoding/json"
)

const (
	SECTIONADMIN  = "admin"
	SECTIONPUBLIC = "public"
	SECTIONUSER   = "user"
	SECTIONFLASH  = "flash"
	SECTIONSTACK  = "stack"
	SECTIONSEO    = "seo"
	SECTIONFORM   = "form"
	SECTIONFUNC   = "func"

	ERRORNOTFOUNDSESSIONCONTEXT = "[SessionUpdate] не могу обновить сессию, т.к. объекта сессии в контексте не найдено \n"
	ERRORNOTFOUNDSESSIONUSER    = "[SessionUpdate] не могу обновить сессию, т.к. не найден такой кукис в таблице с сессиями \n"
)

var (
	DATASECTION = []string{SECTIONADMIN, SECTIONPUBLIC, SECTIONUSER, SECTIONFLASH,
		SECTIONSEO, SECTIONSTACK, SECTIONFORM, SECTIONFUNC}
)

type Carry struct {
	r      *http.Request
	w      http.ResponseWriter
	mix    Gomixer
	node   *node
	params map[string]string
	data   map[string]map[string]interface{}
	csrf   *CSRF
	flash *Flash
	//static error found
	staticerror bool
	notfound    bool
}

//работа с сессиями
func (c *Carry) SessionGet() *Session {
	session := c.GetContextValue("session")
	if session != nil {
		return session.(*Session)
	}
	return nil
}

//первичная инициализация
func (c *Carry) initDefaultValues() {
	//data
	for _, x := range DATASECTION {
		c.data[x] = make(map[string]interface{})
	}

	//c.csrf.Key = c.csrf.randomGenerate(CSRF_LENGTH_KEY)
	c.csrf.Csrf_form = c.csrf.wrapper(true, false)
	c.csrf.Csrf_head = c.csrf.wrapper(false, true)

	//добавление в контекст внешних инжектов для доступа в рамках контекста
	if len(c.mix.stockdata) > 0 {
		for section, mapper := range c.mix.stockdata {
			//c.Log().Printf("Key: %v Value: %v\n", section, mapper)
			if len(mapper) > 0 {
				for k, v := range mapper {
					c.DataSet(section, k, v)
				}
			}
		}
	}

	//добавление в сессию разных функций + значений этих функций  для анализа в шбалонах
	c.DataSet("func", "staticerror", c.NotFoundError())
	c.DataSet("func", "notfound", c.StaticError())
	c.DataSet("func", "r", c.r)
	c.DataSet("func", "w", c.w)
	c.DataSet("func", "realpath", c.Realpath())
	c.DataSet("func", "path", c.Path())
	c.DataSet("func", "flash", c.FlashGetMessage)
	c.DataSet("func", "flashmsg", c.FlashHaveMessage)
	c.DataSet("func", "csrfhead", c.csrf.Csrf_head)
	c.DataSet("func", "csrfform", c.csrf.Csrf_form)
	c.DataSet("func", "paramget", c.ParamGet)
	c.DataSet("func", "queryget", c.QueryGet)
	c.DataSet("func", "mix", c.mix)


}
func (c *Carry) Convert() *convert {
	return c.mix.Convert
}
func (c *Carry) Transliter() *transliter {
	return c.mix.Transliter
}

//работа с типом `форма`
func (c *Carry) Form() *Form {
	return c.mix.Form
}
func (c *Carry) FormInit(form interface{}) {
	//провожу инциализацию формы
	c.mix.Form.InitForm(form)
	//и заношу сразу в контейнер
	c.DataSet("stack", "form", form)
}
func (c *Carry) FormNewStock(name string) *FormStock{
	return c.mix.Form.NewFormStock(name)
}
func (c *Carry) FormValidate(form interface{}) bool {
	return c.mix.Form.ValidateForm(form, c.r)
}
func (c *Carry) FormUpdate(form, source interface{}) {
	c.mix.Form.UpdateForm(form, source)
}
//контейнер для шаблонов
func (c *Carry) Container() map[string]map[string]interface{} {
	return c.data
}
func (c *Carry) Log() *log.Logger {
	return c.mix.Log
}

//общая поддержка
func (c *Carry) InitCarryValuesFromRequest() {
	c.initDefaultValues()
}
func (c *Carry) SetStaticError(status bool) {
	c.staticerror = status
}
func (c *Carry) SetStatusNotFoundError(status bool) {
	c.notfound = status
}
func (c *Carry) NotFoundError() bool{
	return c.notfound
}
func (c *Carry) StaticError() bool {
	return c.staticerror
}

func (c *Carry) Request() *http.Request {
	return c.r
}
func (c *Carry) ResponseWriter() http.ResponseWriter {
	return c.w
}
func (c *Carry) Redirect(url string) *http.Request {
	http.Redirect(c.w, c.r, url, http.StatusFound)
	return nil
}
func (c *Carry) RedirectNotFound() {
	hf := c.mix.wrapperto404handler(c.mix.NotFoundHandler)
	hf(*c)
	return
}


//обработка контекста
func (c *Carry) WriteHTML(httpcode int, text string) error {
	resp := c.w
	resp.Header().Set(ContentType, TextHTMLCharsetUTF8)
	resp.WriteHeader(httpcode)
	resp.Write([]byte(text))
	return nil
}
//записывает json(byte format) в responseWriter
func (c *Carry) JSONB(httpcode int, b []byte) (error) {
	resp := c.w
	resp.Header().Set(ContentType, ApplicationJavaScriptCharsetUTF8)
	resp.WriteHeader(httpcode)
	resp.Write(b)
	return nil
}
//записывает json в responseWriter
func (c *Carry) JSON(code int, answer interface{}) (err error) {
	b, err := json.Marshal(answer)
	if err != nil {
		c.Log().Printf(err.Error())
		return err
	}
	return c.JSONB(code, b)
}
//обработка страницы
func (c *Carry) RenderPage(name string, data interface{}) error {
	return c.mix.Render.Render(name, data, c.w)
}
func (c *Carry) RenderCode(httpcode int, name string, data interface{}) error {
	return c.mix.Render.RenderCode(httpcode, name, data, c.w)
}
func (c *Carry) RenderTxt(httpcode int, name string) error {
	return c.mix.Render.RenderTxt(httpcode, name, c.w)
}
func (c *Carry) ReloadTemplates() {
	c.mix.Render.ReloadTemplate()
}


//работа с параметрами
func (c *Carry) ParamShow() map[string]string {
	return c.params
}
func (c *Carry) ParamGet(key string) string {
	if v, found := c.params[key]; found {
		return v
	}
	return ""

}

//путь, хост
func (c *Carry) Realpath() string {
	if c.node != nil {
		return c.node.realpath
	}
	return ""
}
func (c *Carry) Path() string {
	if c.node != nil {
		return c.node.path
	}
	return ""
}
func (c *Carry) Host() string {
	return c.r.Host
}
func (c *Carry) Method() string {
	return c.r.Method
}

//параметры URL при GET запросах
func (c *Carry) QueryGet(key string) (result string) {
	return c.r.URL.Query().Get(key)
}
func (c *Carry) QuerySet(key, value string) {
	c.r.URL.Query().Set(key, value)
}
func (c *Carry) QueryAdd(key, value string) {
	c.r.URL.Query().Add(key, value)
}
func (c *Carry) QueryEncode() (result string) {
	return c.r.URL.Query().Encode()
}
//получаю множественное занчение типа слайса из формы
func (c *Carry) FormPostMultiGetValue(key string) ([]string) {
	c.r.ParseForm()
	return c.r.Form[key]
}

//параметры=значения полей форм при POST, PUT методах
func (c *Carry) FormPostGetValue(key string) (result string) {
	err := c.r.ParseMultipartForm(1 << 20)
	if err != nil {
		c.mix.Log.Printf(err.Error())
		return
	}
	return c.r.PostFormValue(key)
}

//получение файла из формы с файловам input
func (c *Carry) FormFile(filename string, sizeBytes int64) (multipart.File, *multipart.FileHeader) {
	err := c.r.ParseMultipartForm(sizeBytes)
	if err != nil {
		c.mix.Log.Printf(err.Error())
		return nil, nil
	}
	f, fheader, err := c.r.FormFile(filename)
	if err != nil {
		c.mix.Log.Printf(err.Error())
		return nil, nil
	}
	return f, fheader
}

//получение файлов из формы с файловам input
func (c *Carry) FormMultiFiles(sizeBytes int64) (map[string][]*multipart.FileHeader) {

	err := c.r.ParseMultipartForm(sizeBytes)
	if err != nil {
		c.mix.Log.Printf(err.Error())
		return nil
	}
	return  c.r.MultipartForm.File
}

//получение значений полей при отправке формы при использовании GET метода
func (c *Carry) FormGetValue(key string) (result string) {
	//err := c.r.ParseMultipartForm(1 << 20)
	err := c.r.ParseForm()
	if err != nil {
		c.mix.Log.Printf(err.Error())
		return
	}
	return c.r.Form.Get(key)
}

//асинхронная загрузка файлов по указанному пути
func (c *Carry) UploadSingleFile(formFileName string, sizeBytes int64, pathtoSave string) error {
	f, header := c.FormFile(formFileName, sizeBytes)
	if f == nil && header == nil {
		return errors.New("файл не найден в форме для загрузки")
	}
	go c.uploadfile(*header, f, pathtoSave)
	return nil
}
func (c *Carry) UploadMultiFiles(sizeBytes int64, pathtosaveFiles string, nameUploadForm string) error {
	fm := c.FormMultiFiles(sizeBytes)
	if fm == nil {
		return errors.New("файлов в переданной форме не найдено")
	}
	files := fm[nameUploadForm]
	//обработка списка полученных файлов
	for _, f := range files {
			ff, err := f.Open()
			if err != nil {
				c.Log().Printf(err.Error())
				return err
			}
			//асинхроная загрузка файла по указанному пути
			go c.uploadfile(*f, ff, pathtosaveFiles)
	}
	return nil
}
//загрузка одиночного файла полученного из формы,открытого для чтения,по указанному пути
//используется внутри модуля как горутина для асинхронной загрузки множества
func (c *Carry) uploadfile(header multipart.FileHeader, f multipart.File, pathSaveFile string) {
	defer f.Close()
	dst, err := os.Create(pathSaveFile + header.Filename)
	if err != nil {
		c.Log().Printf(err.Error())
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, f); err != nil {
		c.Log().Printf(err.Error())
		return
	}
	c.Log().Printf("файл `%s` успешно загружен по пути `%s`\n", header.Filename, pathSaveFile)
	return
}

//работа с сессией
func (c *Carry) DataSet(section, key string, value interface{}) {
	_, found := c.data[section]
	if found {
		c.data[section][key] = value
	} else {
		c.data[section] = make(map[string]interface{})
		c.data[section][key] = value
	}
}
func (c *Carry) DataGet(section, key string) (interface{}) {
	sec, found := c.data[section]
	if found {
		value, ok := sec[key]
		if ok {
			return value
		}
	}
	return nil
}

//работа с контекстом
func (c *Carry) SetContextValue(key string, value interface{}) {
	ctx := context.WithValue(c.r.Context(), key, value)
	c.r = c.r.WithContext(ctx)
}
func (c *Carry) GetContextValue(key string) (value interface{}) {
	return c.r.Context().Value(key)
}

//работа с флешем
func (c *Carry) FlashAddMessage(status, section string, message interface{}) {
	c.flash.Set(status, section, message)
}
func (c *Carry) FlashGetMessage(section string) *FlashMessage {
	return c.flash.Get(section)
}
func (c *Carry) FlashHaveMessage(section string) bool {
	return c.flash.HaveMsg(section)
}

//работа с CSRF
func (c *Carry) CSRFVeryfyToken() bool {
	return c.csrf.VerifyToken(c)
}
func (c *Carry) CSRFVeryfyTokenString(key string) bool {
	return c.csrf.VerifyTokenString(key)
}
//работа с куками
func (c *Carry) GetCook(cookName string) *http.Cookie {
	cook, err := c.r.Cookie(cookName)
	if err != nil {
		c.Log().Printf(err.Error())
		return nil
	}
	return cook
}
func (c *Carry) NewCook(cookName string, salt string) http.Cookie {
	return c.mix.NewCook(cookName, salt, *c)
}
func (c *Carry) GenerateCookValue() string {
	return c.mix.cookgeneratenew(defaultCookieSalt)
}
func (c *Carry) SetCookieString(cookValue string, cookName string) {
	cook := c.mix.NewCook(cookName, "", *c)
	cook.Value = cookValue
	http.SetCookie(c.w, &cook)
}
func (c *Carry) SetCookie(cook http.Cookie) bool {
	http.SetCookie(c.w, &cook)
	return true
}
func (c *Carry) StaticFileRender(realpath string) error {
	http.ServeFile(c.w, c.r, realpath)
	return nil
}