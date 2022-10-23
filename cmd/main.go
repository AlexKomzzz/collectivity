package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/AlexKomzzz/collectivity/pkg/handler"
	"github.com/AlexKomzzz/collectivity/pkg/repository"
	"github.com/AlexKomzzz/collectivity/pkg/service"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

func initConfig() error { //Инициализация конфигураций
	viper.AddConfigPath("configs")
	viper.SetConfigName("config")
	return viper.ReadInConfig()
}

func init() {
	// Инициализируем конфигурации
	if err := initConfig(); err != nil {
		logrus.Fatalln("error initializing configs: ", err)
		return
	}

	// загрузка переменных окружения из файла .env
	if err := godotenv.Load(); err != nil { //Загрузка переменного окружения (для передачи пароля из файла .env)
		logrus.Fatalf("error loading env variables: %s", err.Error())
		return
	}
}

func main() {

	db, err := repository.NewPostgresDB(repository.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.username"),
		Password: viper.GetString("db.password"),
		DBName:   viper.GetString("db.dbname"),
	})
	if err != nil {
		logrus.Fatalln("failed to initialize db: ", err)
		return
	}
	defer db.Close()

	repos := repository.NewRepository(db)
	service := service.NewService(repos)
	// handler := handler.NewHandler(service, handler.NewWebClient(make(map[string][]*websocket.Conn), context.Background()))
	handler := handler.NewHandler(service)

	server, err := handler.InitRoutes()
	if err != nil {
		logrus.Fatalf("Error init server: %s", err.Error())
		return
	}

	go func() {
		if err := server.Run(viper.GetString("port")); err != nil {
			logrus.Fatalf("Error run web serv")
			return
		}
	}()

	logrus.Print("Server Started")

	// остановка сервера
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logrus.Print("Server Stopted")

	if err := db.Close(); err != nil {
		logrus.Fatalf("error occured on db connection close: %s\n", err.Error())
	}
}
