package pkg

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

func TestCreateDummyHash(t *testing.T) {
	randomStr := []byte("kFDfSTnuYZq3xnt-HeDmibGNPBIgwzg")
	hashByte, err  := bcrypt.GenerateFromPassword(randomStr, bcrypt.DefaultCost)
	assert.NoError(t, err)
	fmt.Println(string(hashByte))
}