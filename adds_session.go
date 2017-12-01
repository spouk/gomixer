package gomixer

import (
	"time"
	"fmt"
	"crypto/md5"
	"log"
	"io"
	"net/http"
)

const (
	defaultSessionSalt = "e345FG345DFG$%@#$%dsfg"
	logSessionPrefix   = "[session-log] "
	logSessionFlags    = log.Ldate | log.Ltime | log.Lshortfile
)

type Session struct {
	timer            *time.Ticker
	timerTime        time.Duration
	timerFuncSession func() error
	m                *Gomixer
	cookiename       string
	logger           *log.Logger
	sessionTime      time.Duration

	//дополнительные приблуды для удобства работы
	Transliter *transliter
	Convert    *convert
	StateData  *statelessData
	Form       *Form
	Flash      *Flash
}

//timerFuncsession - любая функцию, которая будет вызываться каждый timerSession период ,
//эта функция может быть проверка истекших сессий или еще чего
//эта функция будет дергаться как горутина
func NewSession(m *Gomixer, cookieName string, logout io.Writer,
	sessionTime time.Duration, timerTime time.Duration, timerSession bool, timerFuncSession func() (error)) *Session {
	s := &Session{
		m:                m,
		cookiename:       cookieName,
		logger:           log.New(logout, logSessionPrefix, logSessionFlags),
		timer:            time.NewTicker(timerTime), //время выполнения функции проверки истекшей сессии etc...
		timerTime:        timerTime,                 //время тика таймера
		sessionTime:      sessionTime,               //время самой сессии
		timerFuncSession: timerFuncSession,
		Transliter:       newTransliter(),
		StateData:        newstatelessData(),
		Form:             NewForm(),
	}
	//добавляю логгер конвертеру
	s.Convert = newConverter(s.logger)
	//загружаю деволтные значения в форму
	s.Form.LoadFieldForms(defforms)
	//связываю текущую сессию с мультиплексором для дампа при выходе/сигналах etc...
	//m.session = s
	//запускаю горутину для проверки с периодичностью истекших сессий по таймеру
	go s.sessionTimerRuner()
	//возвращаю инстанс
	return s
}

//горутина, что обрабатывает фукнции переданные в сессию
func (s *Session) sessionTimerRuner() {
	s.logger.Printf("sessionRuner стартанул....\n")
	defer func() {
		s.logger.Printf("sessionRuner закончил\n")
	}()
	for {
		select {
		case <-s.timer.C:
			if s.timerFuncSession != nil {
				err := s.timerFuncSession()
				if err != nil {
					s.logger.Printf(err.Error())
				}
			}
		default:
			time.Sleep(time.Second * 2)
		}
	}
}

//создаю новый кукис
func (s *Session) NewCook(c Carry) (http.Cookie) {
	cook := http.Cookie{}
	cook.Name = s.cookiename
	cook.Value = s.cookgeneratenew(defaultSessionSalt)
	cook.Expires = time.Now().Add(time.Duration(86000*30) * time.Minute)
	cook.Path = "/"
	return cook
}

//установка кукиса
func (s *Session) SetCookie(newcook string, c Carry) bool {
	ncook := s.NewCook(c)
	ncook.Value = newcook
	http.SetCookie(c.w, &ncook)
	return true
}

//генерация нового значения для кукиса
func (s *Session) cookgeneratenew(salt string) string {
	if len(salt) == 0 {
		return fmt.Sprintf("%x", md5.Sum([]byte(time.Now().String()+defaultSessionSalt)))
	} else {
		return fmt.Sprintf("%x", md5.Sum([]byte(time.Now().String()+salt)))
	}
}

//работа с флешем

////проверка сессии на наличие как активной
//func (s *Session) foundExpiredSessionCooks() {
//	defer func() {
//		s.logger.Printf("`foundExpiredSessionCooks` закончил...\n")
//	}()
//	s.logger.Printf("`foundExpiredSessionCooks` стартанул...\n")
//	for {
//		s.Lock()
//		for _, user := range s.Stock {
//			//lastconnect  = время в unix.time последнего подключения
//			//periodSession = время сессии в формате unix.time
//			//идет сравнение lastconnect + periodSession с текущим временем если текущее больше значит сессии истекла
//			//s.logger.Printf("Lastconnect: %v\n+SessionTime: %v\niTimeNow: %v\n",
//			//	time.Unix(user.Lastconnect, 60).String(),
//			//	time.Unix(user.Lastconnect, 60).Add(s.periodSession).String(),
//			//	time.Now().String(),
//			//)
//			if time.Unix(user.Lastconnect, 60).Add(s.periodSession).Unix() < time.Now().Unix() && user.Logged {
//				//s.logger.Printf("FOUND EXPIRED SESSION\n")
//				user.Logged = false
//			}
//			//if user.Logged && ((user.Lastconnect + int64(s.periodSession)) < time.Now().Unix()) {
//			//	user.Logged = false
//			//	s.logger.Printf("Found expired session: %v\n", user.Cookie)
//			//}
//			//if user.Logged && ((user.Lastconnect + int64(s.periodSession)) < time.Now().Unix()) {
//			//	user.Logged = false
//			//	s.logger.Printf("Found expired session: %v\n", user.Cookie)
//			//}
//		}
//		s.Unlock()
//		s.logger.Printf("`foundExpiredSessionCooks` иду на боковую...\n")
//		time.Sleep(s.checkexpiredSession)
//	}
//}

func (s *Session) SessionMiddleware(h HandlerFunc) HandlerFunc {
	return HandlerFunc(func(c Carry) error {
		//добавляю в текущий контекст объект сессии
		c.SetContextValue("session", s)
		c.DataSet("stack", "session", s)

		//if err != nil {
		//	//кукис не найден, создаю новый
		//	s.logger.Printf(err.Error())
		//	newuser, newcook := s.newuser(c)
		//	c.DataSet("user", "session", newuser)
		//	//устанавливаю кукис
		//	http.SetCookie(c.w, &newcook)
		//
		//} else {
		//	//кукис найден, пробую получить из стека
		//	user, found := s.Stock[cook.Value]
		//	if found {
		//		//кукис найден, обновляю данные по текущему запросу
		//		//обновляю последнее время соединения
		//		user.Lastconnect = time.Now().Unix()
		//		//устаналиваю в контекст текущую сессию
		//		c.DataSet("user", "session", user)
		//	} else {
		//		//кукис не найден, создаю новоый кукис и устанавливаю его
		//		newuser, newcook := s.newuser(c)
		//		c.DataSet("user", "session", newuser)
		//		//устанавливаю кукис
		//		http.SetCookie(c.w, &newcook)
		//	}
		//}
		////
		////c.Log().Printf(">>>>>SESSION ENABLED<<<<<<\n")
		////запускаю обработчик запроса
		h(c)
		return nil
	})
}
