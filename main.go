package main

import (
	"database/sql"
	"fmt"
	ioutil "io"
	"log"
	stdHttp "net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/keploy/go-sdk/integrations/kgin/v1"
	"github.com/keploy/go-sdk/integrations/khttpclient"
	"github.com/keploy/go-sdk/integrations/ksql/v1"
	"github.com/keploy/go-sdk/keploy"
	"github.com/lib/pq"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func setupRouter(r *gin.Engine, db *gorm.DB, interceptor *khttpclient.Interceptor) *gin.Engine {
	// Test http call
	r.POST("/test_http_and_sql", func(c *gin.Context) {

		interceptor.SetContext(c.Request.Context())

		client := stdHttp.Client{
			Transport: interceptor,
		}

		req, err := stdHttp.NewRequest("GET", "https://catfact.ninja/fact", nil)
		if err != nil {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
			return
		}

		resp, err := client.Do(req)
		if err != nil {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
			return
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			c.JSON(500, gin.H{
				"message": err.Error(),
			})
			return
		}

		txn := db.WithContext(c.Request.Context()).Exec("SELECT 1")
		if txn.Error != nil {
			c.JSON(500, gin.H{
				"message": txn.Error.Error(),
			})
			return
		}

		c.Data(200, "application/json", body)
	})

	return r
}

func main() {

	dbHost := "localhost"
	dbPort := "5438"
	dbUser := "postgres"
	dbPassword := "postgres"
	dbName := "postgres"

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Warn, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      false,       // Don't include params in the SQL log
			Colorful:                  true,
		},
	)

	driver := ksql.Driver{Driver: &pq.Driver{}}

	sql.Register("keploy", &driver)

	db, err := gorm.Open(postgres.Dialector{
		Config: &postgres.Config{
			DriverName: "keploy",
			DSN:        dsn,
		},
	}, &gorm.Config{
		Logger:               newLogger,
		DisableAutomaticPing: true,
	})
	if err != nil {
		log.Fatal("Error: ", err)
	}

	k := keploy.New(keploy.Config{
		App: keploy.AppConfig{
			Name: "test-bug",
			Port: "8080",
			Host: "localhost",
		},
		Server: keploy.ServerConfig{
			URL: "http://localhost:6789/api",
		},
	})

	r := gin.Default()

	kgin.GinV1(k, r)

	interceptor := khttpclient.NewInterceptor(stdHttp.DefaultTransport)

	r = setupRouter(r, db, interceptor)
	err = r.Run(":8080")
	if err != nil {
		log.Fatal("Error: ", err)
	}
}
