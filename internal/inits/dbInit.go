package inits

import (
	"database/sql"

	"github/JustGopher/Gotaxy/internal/global"
	"github/JustGopher/Gotaxy/internal/storage/models"

	_ "modernc.org/sqlite"
)

// DBInit 数据库初始化
func DBInit() *sql.DB {
	var err error
	db, err := sql.Open("sqlite", "data/data.db")
	if err != nil {
		global.Log.Errorf("DBInit() 打开数据库失败 -> %v", err)
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		global.Log.Errorf("DBInit() 数据库连接失败 -> %v", err)
		panic(err)
	}

	err = models.CreateCfgStructure(db)
	if err != nil {
		global.Log.Errorf("DBInit() 创建配置表结构失败 -> %v", err)
		panic(err)
	}
	err = models.CreateMpgStructure(db)
	if err != nil {
		global.Log.Errorf("DBInit() 创建映射表结构失败 -> %v", err)
		panic(err)
	}
	return db
}
