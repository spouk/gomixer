package gomixer

import (
	"strconv"
	"reflect"
	"strings"
	"fmt"
	"log"
	"net/http"
	"time"
)

//---------------------------------------------------------------------------
//  ОПРЕДЕЛЕНИЕ ТИПОВ
//---------------------------------------------------------------------------
type (
	//---------------------------------------------------------------------------
	//  основная структура форм
	//---------------------------------------------------------------------------
	Form struct {
		FieldsForms map[string]FieldForm
	}
	//---------------------------------------------------------------------------
	//  дефолтные значения для формы
	// Placeholders - ключ = название поля формы, значение ключа - значения `placeholder` для поля формы
	// Value - ключ = название поля формы, значение ключа = содержимое формы в виде `reflect.Value`
	// Error - ошибка при обработка формы, выставляется при проверке формы на валидность полученных значений
	//---------------------------------------------------------------------------
	FieldForm struct {
		Placeholder  string      //описание поля
		Value        interface{} //значение поля
		Error        ErrorForm   //структура для ошибки
		SuccesClass  string      //css класс для успешной обработки
		DefaultValue interface{} //дефолтное значение поля при инициализации
	}
	FormStock struct {
		NameForm     string                 //произвольное название формы
		Desc         map[string]string      //стокер для формы .stack.form.Desc.<FieldName>
		Value        map[string]interface{} //значения для формы .stack.form.Desc.<FieldName>
		DefaultValue map[string]interface{} //значения для формы .stack.form.Desc.<FieldName>
		Error        map[string]ErrorForm   // ошибка для формы .stack.form.Error.<FieldName>
		SuccessClass map[string]string      //css класс для успешной обработки
	}
	ErrorForm struct {
		Error      string
		ErrorClass string
	}
)

//---------------------------------------------------------------------------
//  ФУНКЦИОНАЛ
//---------------------------------------------------------------------------
func NewForm() *Form {
	return &Form{
		FieldsForms: make(map[string]FieldForm),
	}
}

//---------------------------------------------------------------------------
//  функции для работы с дефолтными значениями
//---------------------------------------------------------------------------
func (g *Form) NewFormStock(nameform string) *FormStock {
	return &FormStock{
		Desc:         make(map[string]string),
		Value:        make(map[string]interface{}),
		DefaultValue: make(map[string]interface{}),
		Error:        make(map[string]ErrorForm),
		SuccessClass: make(map[string]string),
	}
}
func (g *Form) LoadFieldForm(key string, f FieldForm) {
	g.FieldsForms[key] = f
}
func (g *Form) LoadFieldForms(keys map[string]FieldForm) {
	for key, value := range keys {
		g.FieldsForms[key] = value
	}
}

//---------------------------------------------------------------------------
//  функции для работы с формой непосредственно
//---------------------------------------------------------------------------
//проверка формы на валидность
func (g *Form) checkedForm(form interface{}) (reflect.Type, reflect.Value, *FormStock) {
	//проверка формы на то, чтобы форма была указателем на форму, что позволит изменять значения в форме
	if reflect.ValueOf(form).Kind() != reflect.Ptr {
		log.Fatalf("[goforms][fatal error] -форма должна передаваться по ссылке- `%v`\n", form)
	}

	//инициализация формы
	t := reflect.TypeOf(form).Elem()
	v := reflect.ValueOf(form).Elem()

	//проверка на наличие стокера, если нет то ошибка
	stocker := t.Field(0)

	////получаем элемент структуры
	//f := reflect.Indirect(reflect.ValueOf(obj)).Field(x)
	////получаем имя элемента структуры
	//name := reflect.TypeOf(obj).Elem().Field(x).Name

	if stocker.Name != "Stock" {
		log.Fatalf("[goforms][fatal error] -invalid struct form- `%v`\n", form)
	}
	stock := v.Field(0).Interface().(*FormStock)
	return t, v, stock
}

//обработка единичного поля и загрузка данных в него из данных html формы по типу
func (g *Form) atomicLoadField(t reflect.StructField, v reflect.Value, stocker *FormStock, r *http.Request) {
	r.ParseForm()
	switch v.Kind() {
	case reflect.String:
		//fmt.Printf("[ATOMIC STRING] %v : %v\n", t.Name,r.Form.Get(t.Name))
		stocker.Value[t.Name] = strings.TrimSpace(r.Form.Get(t.Name))
		v.SetString(strings.TrimSpace(r.Form.Get(t.Name)))

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		value, _ := strconv.Atoi(strings.TrimSpace(r.Form.Get(t.Name)))
		stocker.Value[t.Name] = value
		v.SetInt(int64(value))

	case reflect.Float32, reflect.Float64:
		value, _ := strconv.ParseFloat(r.Form.Get(t.Name), 64)
		stocker.Value[t.Name] = value
		v.SetFloat(value)

	case reflect.Bool:
		val := strings.TrimSpace(r.FormValue(t.Name))
		if val != "" {
			v.SetBool(true)
			stocker.Value[t.Name] = true
		} else {
			v.SetBool(false)
			stocker.Value[t.Name] = true
		}

	case reflect.Slice, reflect.Array:
		//fu := g.FieldsForms[strings.ToLower(  t.Name)]
		resultForm := r.Form[t.Name]

		//switch fu.DefaultValue.(type) {
		switch v.Interface().(type) {
		case []float64:
			var res []float64
			for _, x := range resultForm {
				val, _ := strconv.ParseFloat(x, 10)
				res = append(res, val)
			}
			stocker.Value[t.Name] = res
			v.Set(reflect.ValueOf(&res).Elem())

		case []int64:
			var res []int64
			for _, x := range resultForm {
				val, _ := strconv.ParseInt(x, 10, 0)
				res = append(res, val)
			}
			stocker.Value[t.Name] = res
			v.Set(reflect.ValueOf(&res).Elem())

		case []int:
			fmt.Printf("ATOMIC][RESULT][SLICE:INT] resultform: %v: %T\n", resultForm, resultForm)
			var res []int
			for _, x := range resultForm {
				val, _ := strconv.ParseInt(x, 10, 0)
				res = append(res, int(val))
			}
			stocker.Value[t.Name] = res
			v.Set(reflect.ValueOf(&res).Elem())

		case []string:
			var res []string
			for _, d := range resultForm {
				res = append(res, d)
			}
			stocker.Value[t.Name] = res
			v.Set(reflect.ValueOf(&res).Elem())
		}

	default:
		fmt.Printf("[goforms][atomicLoadField][DEFAULT] %T\n", v)
	}
	return
}

//инициализация формы, вызывается всегда первой, когда используется форма
func (g *Form) InitForm(form interface{}) {
	//проверка формы
	t, v, stock := g.checkedForm(form)

	//должна быть структура, т.к. идет обработка формы, иначе вылет по фатальной ошибке
	switch v.Kind() {
	case reflect.Struct:
		for x := 0; x < v.NumField(); x ++ {
			//добавить дефолтные значения для полей формы
			fu, found := g.FieldsForms[strings.ToLower(t.Field(x).Name)]
			if found {
				stock.Desc[t.Field(x).Name] = fu.Placeholder
				//stock.Error[t.Field(x).Name] = fu.Error
				if fu.DefaultValue != nil {
					stock.DefaultValue[t.Field(x).Name] = fu.DefaultValue
				} else {
					stock.DefaultValue[t.Field(x).Name] = nil
				}
			} else {
				//log.Printf("[goforms][error][key: %s] `u,found:= g.FieldsForms[strings.ToLower(t.Field(x).Name)]` = пустая, нет дефолтных значений\n",strings.ToLower(t.Field(x).Name))
			}
		}
	default:
		log.Fatalf("[goforms][fatal error] -invalid struct form- `%v`\n", form)
	}
}

//загрузка данных из HTML формы в пользовательскую форму, при отправке формы на сервер
func (g *Form) LoadForm(form interface{}, r *http.Request) {
	//проверка формы
	t, v, stock := g.checkedForm(form)
	r.ParseForm()

	for x := 0; x < v.NumField(); x ++ {
		//f := reflect.Indirect(reflect.ValueOf(form)).Field(x)
		//name := reflect.TypeOf(form).Elem().Field(x).Name
		//ff := reflect.TypeOf(form).Elem().Field(x)
		//vv := reflect.Indirect(reflect.ValueOf(&form)).Field(x)
		//fmt.Printf("----CANADDR: %v\n", vv.CanAddr())

		vf := v.Field(x)
		tf := t.Field(x)
		//g.atomicLoadField(tf, vf, stock, r)
		g.atomicLoadField(tf, vf, stock, r)
	}
	return
}

//обновление пользовательской формы данными из формы-источника(из базы данных как пример)
func (g *Form) UpdateForm(form, source interface{}) {
	//проверка формы
	t, v, stock := g.checkedForm(form)

	//проверка объекта источника на то, чтобы форма была указателем на форму, что позволит изменять значения в форме
	if reflect.ValueOf(source).Kind() != reflect.Ptr {
		log.Fatalf("[goforms][fatal error] -форма источник должна передаваться по ссылке- `%v`\n", source)
	}

	//загрузка и обработка источника
	sv := reflect.ValueOf(source).Elem()
	st := reflect.TypeOf(source).Elem()

	fmt.Printf("[SOURCE] Count: %v\n", sv.NumField())

	//бежим по списку элементов источника
	for i := 0; i < sv.NumField(); i++ {
		vf := sv.Field(i)
		tf := st.Field(i)
		//fmt.Printf("[updateform][SOURCE_ELEMENTS] %v : %v\n", tf.Name, vf.Interface())
		//получаю фиелд из формы по имени
		//fmt.Printf("DEBUG-UPDATEFORM: %v\n", tf.Name)
		if tform, found := t.FieldByName(tf.Name); found {

			//типы совпадают, можно делать перенос данных
			if tform.Type == tf.Type {
				//получаю объект значение из формы по имени элемента
				tformValue := v.FieldByName(tf.Name)
				//fmt.Printf("[updateform] Form:[%v:%v] Source:[%v:%v] Source_interface:[%v]\n",
				//	tform.Name, tform.Type, tf.Name, tf.Type, vf.Interface())
				switch vf.Kind() {

				case reflect.String:
					result := strings.TrimSpace(vf.Interface().(string))
					stock.Value[tform.Name] = result
					tformValue.SetString(result)

				case reflect.Int:
					result := vf.Interface().(int)
					stock.Value[tform.Name] = result
					tformValue.Set(reflect.ValueOf(result))

				case reflect.Int8:
					result := vf.Interface().(int8)
					stock.Value[tform.Name] = result
					tformValue.Set(reflect.ValueOf(result))

				case reflect.Int16:
					result := vf.Interface().(int16)
					stock.Value[tform.Name] = result
					tformValue.Set(reflect.ValueOf(result))

				case reflect.Int32:
					result := vf.Interface().(int32)
					stock.Value[tform.Name] = int32(result)
					tformValue.Set(reflect.ValueOf(result))

				case reflect.Int64:
					//fmt.Printf("[int64] [???] [goforms][UpdateForm][error] ===> типы не совпадают [%v]:[%v] :[%v] \n", tform.Type, tf.Type, tf.Name)
					if tf.Type.String() ==  "time.Duration" {
						//fmt.Printf("~!!! FOUND TIME DURATION TYPE  `%v`\n", vf.Interface().(time.Duration).Minutes())
						result := vf.Interface().(time.Duration).Minutes() //конвертирую в минуты в float64
						stock.Value[tform.Name] = result
						tformValue.SetInt(int64(result))

					} else {
						result := vf.Interface().(int64)
						stock.Value[tform.Name] = result
						tformValue.SetInt(result)
					}

				case reflect.Float32:
					result := vf.Interface().(float32)
					stock.Value[tform.Name] = float32(result)
					tformValue.Set(reflect.ValueOf(result))

				case reflect.Float64:
					result := vf.Interface().(float64)
					stock.Value[tform.Name] = float64(result)
					tformValue.Set(reflect.ValueOf(result))

				case reflect.Bool:
					result := vf.Interface().(bool)
					if result == false {
						stock.Value[tform.Name] = false
						tformValue.SetBool(false)
					} else {
						stock.Value[tform.Name] = true
						tformValue.SetBool(true)
					}

				case reflect.Slice, reflect.Array:
					value := vf.Interface()
					switch value.(type) {
					case []int64:
						result := value.([]int64)
						stock.DefaultValue[tform.Name] = result
						stock.Value[tform.Name] = result
						tformValue.Set(reflect.ValueOf(&result).Elem())

					case []string:
						result := value.([]string)
						stock.DefaultValue[tform.Name] = result
						stock.Value[tform.Name] = result
						tformValue.Set(reflect.ValueOf(&result).Elem())

					case []int:
						result := value.([]int)
						stock.DefaultValue[tform.Name] = result
						stock.Value[tform.Name] = result
						tformValue.Set(reflect.ValueOf(&result).Elem())
					}

				case reflect.Invalid:
					fmt.Printf("===>invalid [%T]\n", v)

				default:
					//stock.DefaultValue[tform.Name] = result
					//stock.Value[tform.Name] = result
					//tformValue.Set(reflect.ValueOf(&result).Elem())

					fmt.Printf("[goforms][UpdateForm][default] ===> DEFAULT [%T]\n", v)
				}
			} else {
				//fmt.Printf("[goforms][UpdateForm][error] ===> типы не совпадают [%v]:[%v] :[%v] \n",
				//	tform.Type, tf.Type, tf.Name)
					if tf.Type.String() == "time.Duration" {
						//fmt.Printf("FOUND TIME DURATION TYPE in TFORM.TYPE\n")
						result := vf.Interface().(time.Duration).Minutes() //конвертирую в минуты в float64
						stock.Value[tform.Name] = result
						tformValue := v.FieldByName(tf.Name)
						tformValue.SetInt(int64(result))
					}
					//switch tform.Type.Align().(type) {
					//case time.Duration:
					//	fmt.Printf("FOUND TIME DURATION TYPE in TFORM.TYPE\n")
					//default:
					//	fmt.Printf("UNKNOW TYPE TFORM TYPE: %v\n", tform.Type)
					//}
			}

		} else {
			fmt.Printf("[goforms][UpdateForm][NOT FOUND] %v : %v\n", tf.Name, vf.String())
		}
	}
	return
}

//валидация пользоввательской формы
func (g *Form) ValidateForm(form interface{}, r *http.Request) (status bool) {
	//загрузка данных из HTML формы
	g.LoadForm(form, r)
	//log.Printf("[ValidateForm] form--> %v\n", form)

	//запуск механизма получение данных из формы
	r.ParseForm()

	//проверка формы
	t, v, stock := g.checkedForm(form)

	//количество флагов = количество полей в форме
	var total int = v.NumField() - 1
	var countValidate int = 0

	for i := 0; i < v.NumField(); i ++ {

		//если нет флага, проверка
		if t.Field(i).Tag.Get("form") == "" {

			//получаю дефолтные значения для поля
			fu := g.FieldsForms[strings.ToLower(t.Field(i).Name)]

			//получаю элемент формы
			vf := v.Field(i)
			//fmt.Printf("ValidateForm] %v : %v\n",  t.Field(i).Name, v.Interface())
			switch vf.Kind() {
			default:
				fmt.Printf("[goforms][validateform][ALERT] непроверяемый тип Name: `%v` Value: %v\n", t.Field(i).Name, v.Interface())

			case reflect.Float64, reflect.Float32:
				//result, _ := strconv.ParseFloat(r.Form.Get(t.Field(i).Name), 64)
				result := vf.Interface().(float64)
				if result == 0 {
					//error
					stock.Error[t.Field(i).Name] = fu.Error
				} else {
					stock.Error[t.Field(i).Name] = ErrorForm{}
					stock.SuccessClass[t.Field(i).Name] = fu.SuccesClass
					countValidate ++
				}

			case reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int:
				//result := strings.TrimSpace(r.Form.Get(t.Field(i).Name))
				result := vf.Interface().(int64)
				if result == 0 {
					//error
					stock.Error[t.Field(i).Name] = fu.Error
				} else {
					stock.Error[t.Field(i).Name] = ErrorForm{}
					stock.SuccessClass[t.Field(i).Name] = fu.SuccesClass
					countValidate ++
				}

			case reflect.Slice, reflect.Array:
				//разбор по типу списка
				//resultForm := r.Form[(t.Field(i).Name)]
				value := vf.Interface()
				switch value.(type) {
				case []float64:
					result := value.([]float64)
					if len(result) == 0 {
						//error
						stock.Error[t.Field(i).Name] = fu.Error
					} else {
						stock.Error[t.Field(i).Name] = ErrorForm{}
						stock.SuccessClass[t.Field(i).Name] = fu.SuccesClass
						countValidate ++
					}

				case []int:
					result := value.([]int)
					if len(result) == 0 {
						//error
						stock.Error[t.Field(i).Name] = fu.Error
					} else {
						stock.Error[t.Field(i).Name] = ErrorForm{}
						stock.SuccessClass[t.Field(i).Name] = fu.SuccesClass
						countValidate ++
					}

				case []int64:
					result := value.([]int64)
					fmt.Printf("ValidateForm][slice:int64] %v : %v: result: %v\n", t.Field(i).Name, v.Interface(), result)
					if len(result) == 0 {
						//error
						stock.Error[t.Field(i).Name] = fu.Error
					} else {
						stock.Error[t.Field(i).Name] = ErrorForm{}
						stock.SuccessClass[t.Field(i).Name] = fu.SuccesClass
						countValidate ++
					}

				case []string:
					result := value.([]string)
					if len(result) == 0 {
						//error
						stock.Error[t.Field(i).Name] = fu.Error
					} else {
						stock.Error[t.Field(i).Name] = ErrorForm{}
						stock.SuccessClass[t.Field(i).Name] = fu.SuccesClass
						countValidate ++
					}
				}
			case reflect.String:
				//получаю данные из формы
				//result := strings.TrimSpace(r.Form.Get(t.Field(i).Name))
				result := vf.Interface().(string)
				if result == "" {
					//error
					stock.Error[t.Field(i).Name] = fu.Error
					status = false
				} else {
					stock.Error[t.Field(i).Name] = ErrorForm{}
					stock.SuccessClass[t.Field(i).Name] = fu.SuccesClass
					countValidate ++
				}
			case reflect.Bool:
				//result := r.FormValue(t.Field(i).Name)
				result := vf.Interface().(bool)
				if result == false {
					//error
					stock.Error[t.Field(i).Name] = fu.Error
				} else {
					stock.Error[t.Field(i).Name] = ErrorForm{}
					stock.SuccessClass[t.Field(i).Name] = fu.SuccesClass
					countValidate ++
				}
			}
		} else {
			//флаг есть, вычитем из общего количества
			total --
		}
	}
	//подведение итогов по валидности всей формы
	if total == countValidate {
		fmt.Printf("[goforms][validateform] Total: %v   Numfield: %v   CountValidate: %v , Result: VALIDATE\n", total, v.NumField(), countValidate)
		status = true
	} else {
		status = false
		fmt.Printf("[goforms] [validateform] Total: %v   Numfield: %v   CountValidate: %v , Result: NOT VALIDATE\n", total, v.NumField(), countValidate)
	}
	return
}
