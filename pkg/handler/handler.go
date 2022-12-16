package handler

import (
	"net/http"

	"github.com/AlexKomzzz/collectivity/pkg/service"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *service.Service
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{
		service: services,
	}
}

func (h *Handler) InitRoutes() *gin.Engine { // Инициализация групп функций мультиплексора

	//gin.SetMode(gin.ReleaseMode) // Переключение сервера в режим Релиза из режима Отладка

	mux := gin.New()

	mux.Use(sessions.Sessions("mysession", h.service.NewSession()))
	//mux.LoadHTMLFiles("./web/templates/error.html")
	mux.LoadHTMLGlob("web/templates/*.html")
	mux.NoRoute(Response404) // При неверном URL вызывает ф-ю Response404

	mux.Static("/assets", "./web/assets")
	// mux.StaticFile("/", "index.html")

	// тест
	mux.GET("/test", h.test)

	// пример отправки файла
	// mux.StaticFile("/file", "./web/templates/ex.html")

	// основная страница сайта
	mux.GET("/", h.startList)
	// api := mux.Group("/", h.userIdentity)
	// {
	// 	api.StaticFile("/", "./web/templates/start_list.html")
	// }

	// Стартовая страница для пользователей
	mux.GET("/startList", h.startList)
	mux.POST("/startList", h.startList)

	// работа с файлами
	files := mux.Group("/files")
	{
		// получение файла от клиента
		files.POST("/get-file", h.getFile)
		// отправка данных клиету
		files.GET("/get-data", h.dataClient)
	}

	// телеграм бот
	tlg_bot := mux.Group("/tlg")
	{
		// авторизация
		tlg_bot.GET("/login", h.loginBot)
		// получение данных при авторизации от пользователя
		tlg_bot.POST("/sign-in", h.signInBot)
		// запрос данных
		tlg_bot.POST("/debt", h.getDebtBot)
	}

	// авторизация и аутентификация
	auth := mux.Group("/auth") // Группа аутентификации
	{
		// Войти в систему
		// auth.StaticFile("/login-form", "./web/templates/login.html")
		auth.GET("/login", func(c *gin.Context) {
			c.HTML(http.StatusOK, "login.html", gin.H{})
		})

		// отправка формы для создания пользователя
		// auth.StaticFile("/sign-form", "./web/templates/forma_auth.html")
		auth.GET("/sign-form", func(c *gin.Context) {
			c.HTML(http.StatusOK, "forma_auth.html", gin.H{})
		})

		// создание админа в БД
		auth.GET("/admin", h.createAdm)

		// создание пользователя при получении данных с помощью OAuth (Google или Яндекс)
		auth.POST("/user-oauth", h.createUserOAuth)
		// получение данных при создании нового пользователя, запись в БД регистрации, запрос на подтверждение email
		auth.POST("/sign-up", h.signUp)
		// подтверждение email, проверка токена из URL, добавление пользователя в БД
		auth.GET("/sign-add", h.signAdd)
		// аутентификация пользователя, выдача JWT
		auth.POST("/sign-in", h.signIn)
		// удаление сессии
		auth.GET("/sign-out", h.signOut)

		// восстановление пароля
		pass := auth.Group("/pass")
		{
			// форма восстановления пароля recovery-pass-form
			// auth.StaticFile("/recovery-pass-form", "./web/templates/recovery_pass.html")
			pass.GET("/new-form", func(c *gin.Context) {
				c.HTML(http.StatusOK, "new_pass_email.html", gin.H{})
			})
			// определение пользователя по email
			pass.POST("/definition-user", h.definitionUser)
			// определение пользователя по JWT
			pass.GET("/definition-userJWT", h.definitionUserJWT)
			// восстановление пароля
			pass.POST("/recovery-pass", h.recoveryPass)
		}

		// идентификация через google
		google := auth.Group("/google")
		{
			google.GET("/login", h.oauthGoogleLogin)
			google.GET("/callback", h.oauthGoogleCallback)
		}
		yandex := auth.Group("/yandex")
		{
			yandex.GET("/login", h.oauthYandexLogin)
			yandex.GET("/callback", h.oauthYandexCallback)
		}
	}

	return mux
}
