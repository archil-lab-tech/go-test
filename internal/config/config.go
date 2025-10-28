package config

import (
	"context"
	"os"
	"strings"
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
		MongoURI: strings.TrimSpace(os.Getenv("MONGO_URI")),
		DBName:   getenv("DB_NAME", "app"),
	}
}

// InitMongo tries to connect and ping Atlas, then populates cfg.Mongo and cfg.DB.
func InitMongo(ctx context.Context, cfg *Config) {
	if strings.TrimSpace(cfg.MongoURI) == "" {
		log.Warn().Msg("MONGO_URI not set; Mongo features will be disabled")
		return
	}

	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	opts := options.Client().
		ApplyURI(cfg.MongoURI).
		SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1)).
		SetAppName("go-api").
		SetRetryWrites(true)

	cl, err := mongo.Connect(connectCtx, opts)
	if err != nil {
		log.Error().Err(err).Msg("mongo connect failed")
		return
	}

	pingCtx, cancelPing := context.WithTimeout(ctx, 5*time.Second)
	defer cancelPing()
	if err := cl.Ping(pingCtx, nil); err != nil {
		log.Error().Err(err).Msg("mongo ping failed")
		_ = cl.Disconnect(context.Background())
		return
	}

	cfg.Mongo = cl
	cfg.DB = cl.Database(cfg.DBName)
	log.Info().
		Str("db", cfg.DBName).
		Bool("has_mongo_uri", cfg.MongoURI != "").
		Msg("mongo connected")
}

// EnsureMongo establishes a client if we have a URI but no active client.
func EnsureMongo(ctx context.Context, cfg *Config) {
	if cfg.Mongo == nil && strings.TrimSpace(cfg.MongoURI) != "" {
		InitMongo(ctx, cfg)
	}
}

// CloseMongo closes the client and clears pointers.
func CloseMongo(ctx context.Context, cfg *Config) {
	if cfg.Mongo == nil {
		return
	}
	if err := cfg.Mongo.Disconnect(ctx); err != nil {
		log.Error().Err(err).Msg("mongo disconnect failed")
	} else {
		log.Info().Msg("mongo disconnected")
	}
	cfg.Mongo = nil
	cfg.DB = nil
}

func getenv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}
