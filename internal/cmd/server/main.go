package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Dormant512/all-things-cognitei/internal/config"
	"github.com/Dormant512/all-things-cognitei/internal/handlers"
	"github.com/Dormant512/all-things-cognitei/internal/repository"
	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"log"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	server := gin.Default()
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("cannot parse config\n")
	}

	dsn := cfg.GetConnectionString()

	fmt.Println(dsn)

	db, err := mongo.Connect(ctx, options.Client().ApplyURI(dsn))

	defer func() {
		cancel()
		if err = db.Disconnect(ctx); err != nil {
			log.Fatalf("mongodb disconnect error: %v", err)
		}
	}()

	if err != nil {
		log.Fatalf("connection error: %v", err)
		return
	}

	err = db.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatalf("ping failed: %v", err)
		return
	}
	fmt.Println("ping mongodb successful")

	repo := repository.NewRepository(db, cfg)
	srv := handlers.NewService(repo)

	router := server.Group("/items")
	router.GET("/ping", srv.SrvPing)
	router.GET("/id", srv.SrvGetById)
	router.GET("/category", srv.SrvGetInCategory)

	var dmUsers gin.Accounts
	err = json.Unmarshal([]byte(cfg.DmUserJson), &dmUsers)
	if err != nil {
		log.Fatalf("DM account json parse failed: %v", err)
	}

	authorized := server.Group("/items", gin.BasicAuth(dmUsers))
	authorized.POST("/new", srv.SrvNewItem)
	authorized.DELETE("/delete", srv.SrvDeleteById)
	log.Fatal(server.Run(":" + cfg.AppPort))
}
