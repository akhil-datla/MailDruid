package postgresmanager

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type PostgresManager struct {
	db *gorm.DB
}

var postgresmanager = &PostgresManager{}

func Open(dbname, user, pass string) error {
	var err error
	dsn := fmt.Sprintf("host=localhost dbname=%s port=5432 user=%s password=%s sslmode=disable", dbname, user, pass)
	postgresmanager.db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	return err
}

func AutoCreateStruct(data interface{}) error {
	err := postgresmanager.db.AutoMigrate(data)
	return err
}

func Save(data interface{}) error {
	res := postgresmanager.db.Create(data)
	return res.Error
}

func Query(data, store interface{}) error {
	res := postgresmanager.db.Where(data).First(store)
	return res.Error
}

func Update(model, data interface{}) error {
	res := postgresmanager.db.Model(model).Updates(data)
	return res.Error
}

func Delete(data interface{}) error {
	res := postgresmanager.db.Delete(data)
	return res.Error
}

func ReadAll(store interface{}) {
	postgresmanager.db.Find(store)

}