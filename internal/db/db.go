package db

import (
	"strconv"

	"github.com/kube-orchestra/maestro/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// need to figure out how to use this with gorm
// var DBPool *pgxpool.Pool
var dbClient *gorm.DB

//type ErrorNotFound struct{}

// func (e *ErrorNotFound) Error() string {
// 	return fmt.Sprintf("Resource not found")
// }

func init() {
	println("Initializing DB Client .....")
	dbClient, _ = newClient()
	setupModel()
	println("DB Client Initialing Complete .....")
}

// newClient Creates a DynamoDB Client
func newClient() (*gorm.DB, error) {

	//dsn := "host=somehost user=dbuser password=dbpwd dbname=somename port=5432 sslmode=disable TimeZone=America/Los_Angeles"

	dsn := "host=" + config.Cfg.DBHost + " user=" + config.Cfg.DBUser + " password=" + config.Cfg.DBPass + " dbname=" + config.Cfg.DBName +
		" port=" + strconv.Itoa(config.Cfg.DBPort) + " sslmode=" + config.Cfg.DBSSL + " TimeZone=" + config.Cfg.DBTmz

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		// be careful - this prints out the whole dsn string!!
		println("Unable to connect to database: %v\n", err)
	}
	return db, nil
}
