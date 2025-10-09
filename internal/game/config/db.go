package config

import (
	"Jogo-de-Cartas-Multiplayer-Distribuido/internal/game/domain/entities"
	"fmt"
	"time"
	"os"
	"log"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)



var user = os.Getenv("DB_USER")
var password = os.Getenv("DB_PASSWORD")
var dbname = os.Getenv("DB_NAME")
var host = os.Getenv("DB_HOST")
var port = os.Getenv("DB_PORT")

var dsn = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
    user, password, host, port, dbname)



func CretaeTable() gorm.DB{

	var db *gorm.DB
	var err error

	// Tenta até 10 vezes com 3 segundos de espera entre tentativas
	for i := 1; i <= 10; i++ {
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err == nil {
			break
		}
		log.Printf("[DB] Tentativa %d: aguardando MySQL... (%v)", i, err)
		time.Sleep(3 * time.Second)
	}
	// Cria as tabelas automaticamente
	err = db.AutoMigrate(&entities.Package{} ,&entities.Player{}, &entities.Card{})
	if err != nil {
		panic("Falha ao criar tabelas")
	}
	
	return  *db
}


