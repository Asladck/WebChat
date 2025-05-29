package main

import (
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"time"
	handler "websckt/internal/handler"
	"websckt/internal/repository"
	service "websckt/internal/service"
	ws "websckt/internal/ws_server"
)

const (
	addr = "0.0.0.0:9090"
)

func initConfig() error {
	viper.AddConfigPath("configs")
	viper.SetConfigName("configs")
	return viper.ReadInConfig()
}
func main() {
	logrus.SetFormatter(new(logrus.JSONFormatter))
	if err := initConfig(); err != nil {
		logrus.Fatal("error initializing configs", err)
	}
	if err := godotenv.Load(); err != nil {
		logrus.Fatal("error initializing configs", err)
	}

	db, err := repository.NewPostgresDB(repository.Config{
		Host:     viper.GetString("db.host"),
		Port:     viper.GetString("db.port"),
		Username: viper.GetString("db.username"),
		Password: os.Getenv("DB_PASSWORD"),
		DBName:   viper.GetString("db.dbname"),
		SSLMode:  viper.GetString("db.sslmode"),
	})

	if err != nil {
		logrus.Fatalf("failed to initializate a db: %s", err.Error())
	}
	repos := repository.NewRepository(db)
	services := service.NewService(repos)
	handlers := handler.NewHandler(services)

	wsSrv := ws.NewWsServer(addr)

	router := wsSrv.Engine()
	handlers.InitRouter(router)

	logrus.Info("Started ws server")
	if err := wsSrv.Start(); err != nil {
		logrus.Fatalf("Error with ws server: %v", err)
	}
	time.Sleep(5 * time.Second)
	logrus.Error(wsSrv.Stop())
}
