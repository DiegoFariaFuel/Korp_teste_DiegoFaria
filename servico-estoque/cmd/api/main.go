package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Logger
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// DSN com variável de ambiente
	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		dbHost = "localhost" // fallback local
	}
dbName := os.Getenv("DB_NAME")
if dbName == "" {
    dbName = "faturamento"
}

	dsn := fmt.Sprintf("host=%s user=postgres password=postgres dbname=" + dbName + " port=5432 sslmode=disable", dbHost)
	
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		logger.Fatal("Erro ao conectar ao banco", zap.Error(err))
	}
	db.AutoMigrate(&Produto{})

	// Gin
	r := gin.Default()
	r.Use(cors.Default())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	r.GET("/api/produtos", func(c *gin.Context) {
		var produtos []Produto
		db.Find(&produtos)
		c.JSON(200, produtos)
	})

	r.POST("/api/produtos", func(c *gin.Context) {
		var p Produto
		if err := c.ShouldBindJSON(&p); err == nil {
			db.Create(&p)
			c.JSON(201, p)
		} else {
			c.JSON(400, gin.H{"error": err.Error()})
		}
	})

	r.POST("/api/produtos/reservar", func(c *gin.Context) {
		var req ReservaRequest
		if err := c.ShouldBindJSON(&req); err == nil {
			c.JSON(200, gin.H{
				"reservaId": req.NotaFiscalId,
				"mensagem":  "Reserva simulada com sucesso",
			})
		} else {
			c.JSON(400, gin.H{"error": err.Error()})
		}
	})

	port := "8080"
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	fmt.Println("Serviço estoque rodando na porta", port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Desligando...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Shutdown:", err)
	}
}

type Produto struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Codigo    string `json:"codigo"`
	Descricao string `json:"descricao"`
	Saldo     int    `json:"saldo"`
}

type ReservaRequest struct {
	NotaFiscalId string        `json:"notaFiscalId"`
	Itens        []ReservaItem `json:"itens"`
}

type ReservaItem struct {
	ProdutoId  string `json:"produtoId"`
	Quantidade int    `json:"quantidade"`
}
