package database

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"reflect"
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
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		AllowGlobalUpdate:                        true,
	})
	sqlDb, _ := db.DB()
	sqlDb.SetMaxIdleConns(2)
	sqlDb.SetConnMaxIdleTime(1 * time.Hour)

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

func AddPaginationAndFilter(query string, params []interface{}, offset int, limit int) func(db *gorm.DB) *gorm.DB {

	if query == "" {
		return func(db *gorm.DB) *gorm.DB {
			return db.Scopes(AddPagination(offset, limit))
		}
	}

	return func(db *gorm.DB) *gorm.DB {
		return db.Scopes(AddPagination(offset, limit), AddFilter2(query, params))
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
		//value tiplerine gore sorgu ayarlancak/ ilike, ve json gelcek, < > bunlar da gelcek
		filter.WriteString(fmt.Sprintf(" AND %s = ?", key))
		params = append(params, value)
	}

	return func(db *gorm.DB) *gorm.DB {
		return db.Where(gorm.Expr(filter.String(), params...))
	}

}

func AddFilter2(query string, params []interface{}) func(db *gorm.DB) *gorm.DB {

	return func(db *gorm.DB) *gorm.DB {
		return db.Where(gorm.Expr(query, params...))
	}

}

func InsertSeedData[T any](data []T) error {
	if DB == nil {
		return &DBError{Message: ErrorConnectDB}
	}

	ids := make([]*string, len(data))
	//if data has id field, get all ids
	for i, v := range data {
		val := reflect.ValueOf(v)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		if val.Kind() == reflect.Struct {
			field := val.FieldByName("ID")
			if field.IsValid() {
				fmt.Println(field.String())
				ids[i] = field.Interface().(*string)
			}
		}
	}
	fmt.Println(ids)

	// get exist ids in db
	var existIds []string
	if err := DB.Model(data).Where("id IN ?", ids).Pluck("id", &existIds).Error; err != nil {
		return &DBError{Message: "Seed data eklenirken hata oluştu", Err: err}
	}
	fmt.Println(existIds)

	// remove exist ids from data
	for _, id := range existIds {
		for i, v := range data {
			val := reflect.ValueOf(v)
			if val.Kind() == reflect.Ptr {
				val = val.Elem()
			}
			if val.Kind() == reflect.Struct {
				field := val.FieldByName("ID")
				if field.IsValid() && *field.Interface().(*string) == id {
					data = append(data[:i], data[i+1:]...)
					break
				}
			}
		}
	}
	if len(data) == 0 {
		fmt.Println("No data to insert")
		return nil
	}
	// insert data
	if err := DB.Create(data).Error; err != nil {
		return &DBError{Message: "Seed data eklenirken hata oluştu", Err: err}
	}

	return nil
}

func InsertSeedDataFromSQLFile(db *gorm.DB, path string) error {
	if db == nil {
		return &DBError{Message: ErrorConnectDB}
	}

	sqlFile, err := os.Open(path)
	if err != nil {
		return &DBError{Message: "Seed data dosyası açılamadı",

			Err: err}
	}
	defer sqlFile.Close()

	scanner := bufio.NewScanner(sqlFile)
	sqltext := ""

	if err := scanner.Err(); err != nil {
		return &DBError{Message: "Seed data dosyası okunurken hata oluştu", Err: err}
	}
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		sqltext += line
	}

	if err := db.Exec(sqltext).Error; err != nil {
		return &DBError{Message: "Seed data insert edilirken hata oluştu", Err: err}
	}

	return nil
}
