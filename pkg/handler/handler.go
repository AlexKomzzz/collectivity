package handler

import (
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
	mux.LoadHTMLGlob("./web/templates/*.html")
	mux.NoRoute(Response404) // При неверном URL вызывает ф-ю Response404

	mux.StaticFile("/", "./web/templates/index.html")
	// mux.StaticFile("/", "index.html")

	auth := mux.Group("/auth") // Группа аутентификации
	{
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
		auth.POST("/sign-up", h.signUp)
		auth.POST("/sign-in", h.signIn)
	}

	return mux, nil
}
