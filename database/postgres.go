package database

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/lib/pq"

	"github.com/ismailozdel/core2/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	ErrorConnectDB   = "veritabanına bağlanılamadı: %v"
	ErrorAutoMigrate = "otomatik migrasyon hatası: %v"
	ErrorInvalidEnv  = "gerekli çevre değişkenleri eksik"
	LogDBConnected   = "Veritabanı bağlantısı başarılı"
)

var DB *gorm.DB
var CompanyDB map[string]*gorm.DB

// DBError özel veritabanı hata yapısı
type DBError struct {
	Message string
	Err     error
}

// CompanyDBConfig şirket veritabanı yapılandırması
type host struct {
	Host string
	Port string
}

func (e *DBError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func Connect(config *config.DBConfig) error {
	dsn := config.GetDSN()
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		return &DBError{Message: ErrorConnectDB, Err: err}
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetConnMaxIdleTime(3 * time.Second)
	sqlDB.SetMaxOpenConns(10)
	log.Println(sqlDB.Stats().MaxIdleClosed)
	log.Println(sqlDB.Stats().MaxIdleTimeClosed)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		AllowGlobalUpdate:                        true,
		ConnPool:                                 sqlDB,
	})

	if err != nil {
		return &DBError{Message: ErrorConnectDB, Err: err}
	}
	log.Println(LogDBConnected)
	DB = db
	return nil
}

func GetCompanyDB(companyID string) (*gorm.DB, error) {
	if CompanyDB[companyID] != nil {
		return CompanyDB[companyID], nil
	}

	if err := ConnectCompanyDB(companyID); err != nil {
		return nil, err
	}

	return CompanyDB[companyID], nil
}

func ConnectCompanyDB(companyID string) error {
	// eğer CompanyDB map'i nil ise oluştur
	if CompanyDB == nil {
		CompanyDB = make(map[string]*gorm.DB)
	}

	// eğer companyID zaten varsa tekrar bağlanma
	if CompanyDB[companyID] != nil {
		return nil
	}

	var hostId string
	if err := DB.Table("companies").Select("host_id").Where("id = ?", companyID).Scan(&hostId).Error; err != nil {
		return err
	}

	var h host = host{}
	if err := DB.Table("hosts").Select("host, port").Where("id = ?", hostId).Scan(&h).Error; err != nil {
		return err
	}

	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable", h.Host, "postgres", "postgres", "mikroservis_template", h.Port)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		AllowGlobalUpdate:                        true,
	})

	if err != nil {
		return &DBError{Message: ErrorConnectDB, Err: err}
	}
	CompanyDB[companyID] = db
	return nil
}

func AutoMigrate(models ...interface{}) error {
	if DB == nil {
		return &DBError{Message: ErrorConnectDB}
	}

	// Model yapılarını buraya ekleyin
	if err := DB.AutoMigrate(models...); err != nil {
		return &DBError{Message: ErrorAutoMigrate, Err: err}
	}
	return nil
}

func AddPaginationAndFilter(query map[string]interface{}, offset int, limit int) func(db *gorm.DB) *gorm.DB {

	return func(db *gorm.DB) *gorm.DB {
		return db.Scopes(AddPagination(offset, limit), AddFilter(query))
	}
}

func AddPagination(offset int, limit int) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(offset).Limit(limit)
	}
}

func AddFilter(query map[string]interface{}) func(db *gorm.DB) *gorm.DB {
	var filter strings.Builder
	var params []interface{}
	filter.WriteString("1=1")

	for key, value := range query {
		filter.WriteString(fmt.Sprintf(" AND %s = ?", key))
		params = append(params, value)
	}

	return func(db *gorm.DB) *gorm.DB {
		return db.Where(gorm.Expr(filter.String(), params...))
	}

}
