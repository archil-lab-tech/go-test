package config

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Config struct {
	Port     string
	MongoURI string
	DBName   string

	// runtime
	Mongo *mongo.Client
	DB    *mongo.Database
}

func Load() Config {
	return Config{
		Port:     getenv("PORT", "8080"),
		MongoURI: os.Getenv("MONGO_URI"),
		DBName:   getenv("DB_NAME", "app"),
	}
}

func InitMongo(ctx context.Context, cfg *Config) {
	if cfg.MongoURI == "" {
		log.Warn().Msg("MONGO_URI not set; Mongo features will be disabled")
		return
	}
	cl, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		log.Error().Err(err).Msg("mongo connect failed")
		return
	}
	pctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := cl.Ping(pctx, nil); err != nil {
		log.Error().Err(err).Msg("mongo ping failed")
		_ = cl.Disconnect(context.Background())
		return
	}
	cfg.Mongo = cl
	cfg.DB = cl.Database(cfg.DBName)
	log.Info().Str("db", cfg.DBName).Msg("mongo connected")
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
