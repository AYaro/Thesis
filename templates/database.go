package templates

var Database = `package gen

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mssql"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"gopkg.in/gormigrate.v1"
)


type DB struct {
	db *gorm.DB
}

func NewDB(urlStr string) *DB {
	urlStr:= os.Getenv("DATABASE_URL")
	if urlStr == "" {
		panic("NO DATABASE STRING")
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		panic(err)
	}

	db, err := gorm.Open(u.Scheme, urlStr)
	if err != nil {
		panic(err)
	}
	
	if urlStr == "sqlite3://:memory:" {
		db.DB().SetMaxIdleConns(1)
		db.DB().SetConnMaxLifetime(time.Second * 300)
		db.DB().SetMaxOpenConns(1)
	} else {
		db.DB().SetMaxIdleConns(10)
		db.DB().SetMaxOpenConns(10)
	}
	
	return &DB{db: db}
}

func (d *DB) Query() *gorm.DB {
	return d.db
}

func (d *DB) AutoMigrate() error {
	return AutoMigrate(d.db)
}

func (d *DB) Ping() error {
	return d.db.DB().Ping()
}

func (d *DB) Close() error {
	return d.db.Close()
}
