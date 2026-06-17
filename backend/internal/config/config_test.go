package config

import "testing"

func baseValid() *Config {
	return &Config{
		Env:            "production",
		JWTSecret:      "a-sufficiently-long-secret-value",
		AllowedOrigins: "https://poker.example.com",
		DB:             DBConfig{Host: "db", User: "poker", Name: "poker"},
		Redis:          RedisConfig{Host: "redis"},
	}
}

func TestValidate_Valid(t *testing.T) {
	if err := baseValid().Validate(); err != nil {
		t.Fatalf("expected valid config, got %v", err)
	}
}

func TestValidate_MissingJWTSecret(t *testing.T) {
	c := baseValid()
	c.JWTSecret = ""
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for missing JWT secret")
	}
}

func TestValidate_WeakJWTSecretInProduction(t *testing.T) {
	c := baseValid()
	c.JWTSecret = "your-super-secret-jwt-key-change-in-production"
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for weak JWT secret in production")
	}
}

func TestValidate_ShortJWTSecretInProduction(t *testing.T) {
	c := baseValid()
	c.JWTSecret = "short"
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for short JWT secret in production")
	}
}

func TestValidate_ShortSecretAllowedInDev(t *testing.T) {
	c := baseValid()
	c.Env = "development"
	c.JWTSecret = "short"
	if err := c.Validate(); err != nil {
		t.Fatalf("short secret should be allowed in development, got %v", err)
	}
}

func TestValidate_WildcardOriginRejectedInProduction(t *testing.T) {
	c := baseValid()
	c.AllowedOrigins = "*"
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for wildcard origin in production")
	}
}

func TestValidate_MissingDBAndRedis(t *testing.T) {
	c := baseValid()
	c.DB.Host = ""
	c.Redis.Host = ""
	if err := c.Validate(); err == nil {
		t.Fatal("expected error for missing DB/Redis host")
	}
}
