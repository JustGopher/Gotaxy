package inits

import (
	"database/sql"
	"github/JustGopher/Gotaxy/internal/storage/models"
	"log"

	_ "modernc.org/sqlite"
)

func initDB(db *sql.DB) {
	var err error

	db, err = sql.Open("sqlite", "data/data.db")
	if err != nil {
		log.Fatalf("打开数据库失败 -> %v", err)
		return
	}

	defer func(DB *sql.DB) {
		err := DB.Close()
		if err != nil {
			log.Printf("关闭数据库失败 -> %v", err)
		}
	}(db)

	err = db.Ping()
	if err != nil {
		log.Printf("数据库连接失败 -> %v", err)
		return
	}

	models.CreateConfig()
	models.CreateMapping()
}
