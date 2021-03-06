package main

import (
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"log"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	SQLiteFilePath           string          `json:"sqliteFilePath"`
	ServerHost               string          `json:"serverHost"`
	ServerPort               string          `json:"serverPort"`
	RSAPrivKey               *rsa.PrivateKey `json:"-"` // exclude from json response
	RSAPubKey                string          `json:"rsaPubKey"`
	DecryptsAllowedPerPeriod int             `json:"decryptsAllowedPerPeriod"`
	PeriodInSeconds          int             `json:"periodInSeconds"`
}

// CFG is an exportable single source of truth for config info & secrets
var CFG *Config

func defaultRead(envVar string, defaultVal string) (toSet string) {
	envVal := os.Getenv(envVar)
	if envVal == "" {
		return defaultVal
	}
	return envVal
}

func InitConfig() {
	log.Println("Generating configs...")
	RSAPrivateKeyString := strings.TrimSpace(os.Getenv("RSA_PRIVATE_KEY"))

	goodRSAPrefix := strings.HasPrefix(RSAPrivateKeyString, "-----BEGIN RSA PRIVATE KEY-----") == true
	goodRSASuffix := strings.HasSuffix(RSAPrivateKeyString, "-----END RSA PRIVATE KEY-----") == true

	// Basic sanity check:
	if !goodRSAPrefix || !goodRSASuffix {
		panic("RSA Private Key Invalid")
	}

	log.Println("Validating RSA private key...")
	block, _ := pem.Decode([]byte(RSAPrivateKeyString))
	privKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	HandleErr(err)

	log.Println("Generating corresponding RSA public key...")
	pubASN1, err := x509.MarshalPKIXPublicKey(&privKey.PublicKey)
	HandleErr(err)
	pubBytes := pem.EncodeToMemory(&pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: pubASN1,
	})

	decryptsPerPeriodStr := defaultRead("DECRYPTS_PER_PERIOD", "100")
	decryptsPerPeriodInt, err := strconv.Atoi(decryptsPerPeriodStr)
	HandleErr(err)

	periodInSecondsStr := defaultRead("PERIOD_IN_SECONDS", "600")
	periodInSecondsInt, err := strconv.Atoi(periodInSecondsStr)
	HandleErr(err)

	cfg := Config{
		SQLiteFilePath:           defaultRead("SQLITE_FILEPATH", "opsep.sqlite3"),
		ServerHost:               os.Getenv("SERVER_HOST"), // will return "" (good) if not set
		ServerPort:               defaultRead("PORT", "8080"),
		RSAPrivKey:               privKey,
		RSAPubKey:                string(pubBytes),
		DecryptsAllowedPerPeriod: decryptsPerPeriodInt,
		PeriodInSeconds:          periodInSecondsInt,
	}

	// Log all configs except for RSA private key (sensitive):
	b, err := json.Marshal(cfg)
	HandleErr(err)
	log.Println(string(b))

	CFG = &cfg
}
