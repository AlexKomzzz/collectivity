package handler

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/AlexKomzzz/collectivity/pkg/service"
	"github.com/gin-gonic/gin"
)

//go:embed web/assets/* web/templates/*
var f embed.FS

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

	//mux.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler)) // Для работы сваггера

	// Следующий блок кода отвечает за загрузку шаблонов html и css из директории FS
	templ := template.Must(template.New("").ParseFS(f, "web/templates/*.html"))
	fsys, err := fs.Sub(f, "web/assets")
	if err != nil {
		return mux, err
	}
	mux.StaticFS("/assets", http.FS(fsys))
	mux.SetHTMLTemplate(templ)

	mux.NoRoute(Response404) // При неверном URL вызывает ф-ю Response404

	mux.StaticFile("/", "./pkg/handler/web/templates/index.html")
	auth := mux.Group("/auth") // Группа аутентификации
	{
		// идентификация через google
		google := auth.Group("/google")
		{
			google.GET("/login", h.oauthGoogleLogin)
			google.GET("/callback", h.oauthGoogleCallback)
		}
		auth.POST("/sign-up", h.signUp)
		auth.POST("/sign-in", h.signIn)
	}

	// api := mux.Group("/api", h.userIdentity) //Группа для взаимодействия с List
	// {

	// }
	return mux, nil
}
