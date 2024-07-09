package auth

import (
	cryptorand "crypto/rand"
	"github.com/google/uuid"
	"math"
	"math/big"
	"math/rand"
	"strings"
)

func GenerateState() string {
	return uuid.New().String()
}

const Charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ1234567890"

func GenerateSessionIdWithLength(length int) (string, error) {
	source, err := cryptorand.Int(cryptorand.Reader, big.NewInt(int64(math.MaxInt64)))
	if err != nil {
		return "", err
	}

	prng := rand.New(rand.NewSource(source.Int64()))
	builder := strings.Builder{}
	charSetLen := len(Charset)

	for i := 0; i < length; i++ {
		builder.WriteString(string(Charset[prng.Intn(charSetLen)]))
	}
	return builder.String(), nil
}

func GenerateSessionId() (string, error) {
	return GenerateSessionIdWithLength(32)
}

type ResultErrorMessage struct {
	Error string `json:"error,omitempty"`
}

func NewResultErrorMessage(err error) *ResultErrorMessage {
	return &ResultErrorMessage{
		Error: err.Error(),
	}
}
