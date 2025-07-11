package inits

import (
	"database/sql"
	"github/JustGopher/Gotaxy/internal/storage/models"
	"log"

	_ "modernc.org/sqlite"
)

// DBInit 数据库初始化
func DBInit(db *sql.DB) *sql.DB {
	var err error

	db, err = sql.Open("sqlite", "data/data.db")
	if err != nil {
		log.Fatalf("打开数据库失败 -> %v", err)
		return nil
	}

	err = db.Ping()
	if err != nil {
		log.Printf("数据库连接失败 -> %v", err)
		panic(err)
	}

	err = models.CreateCfgStructure(db)
	if err != nil {
		log.Printf("创建配置表结构失败 -> %v", err)
		return nil
	}

	err = models.CreateMpgStructure(db)
	if err != nil {
		log.Printf("创建映射表结构失败 -> %v", err)
		return nil
	}
	return db
}
