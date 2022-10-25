package handler

import (
	"net/http"

	"github.com/AlexKomzzz/collectivity/pkg/service"
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

func (h *Handler) InitRoutes() (*gin.Engine, error) { // Инициализация групп функций мультиплексора

	//gin.SetMode(gin.ReleaseMode) // Переключение сервера в режим Релиза из режима Отладка

	mux := gin.New()

	//mux.LoadHTMLFiles("./web/templates/error.html")
	mux.LoadHTMLGlob("web/templates/*.html")
	mux.NoRoute(Response404) // При неверном URL вызывает ф-ю Response404

	mux.Static("/assets", "./web/assets")
	// mux.StaticFile("/", "index.html")

	mux.GET("/test", h.test)

	// основная страница сайта
	api := mux.Group("/", h.userIdentity)
	{
		api.StaticFile("/", "./web/templates/start_list.html")
	}

	// авторизация и аутентификация
	auth := mux.Group("/auth") // Группа аутентификации
	{
		// Войти в систему
		// auth.StaticFile("/login-form", "./web/templates/login.html")
		auth.GET("/login-form", func(c *gin.Context) {
			c.HTML(http.StatusBadRequest, "login.html", gin.H{})
		})

		// отправка формы для создания пользователя
		//auth.StaticFile("/sign-form/", "./web/templates/forma_auth.html")
		auth.GET("/sign-form", func(c *gin.Context) {
			c.HTML(http.StatusBadRequest, "forma_auth.html", gin.H{})
		})
		auth.POST("/sign-up", h.signUp)
		auth.POST("/sign-in", h.signIn)
		// сброс пароля
		auth.GET("/refresh-pass", h.test)

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

	return mux, nil
}
