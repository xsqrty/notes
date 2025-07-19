package passwd

import "golang.org/x/crypto/bcrypt"

// PasswordGenerator defines methods for generating and comparing hashed passwords.
type PasswordGenerator interface {
	Generate(string) (string, error)
	Compare(hash, password string) bool
}

// passwordGenerator is a struct that encapsulates the logic for generating and comparing password hashes.
// It uses a cost parameter to configure the hashing strength.
type passwordGenerator struct {
	cost int
}

// NewPasswordGenerator creates and returns an instance of PasswordGenerator with the specified bcrypt cost.
func NewPasswordGenerator(cost int) PasswordGenerator {
	return &passwordGenerator{cost: cost}
}

// Generate creates a bcrypt hash from the provided password using the specified cost parameter in the password generator.
func (pg *passwordGenerator) Generate(password string) (string, error) {
	bts, err := bcrypt.GenerateFromPassword([]byte(password), pg.cost)
	if err != nil {
		return "", err
	}

	return string(bts), nil
}

// Compare checks if the provided password matches the given hashed password using bcrypt. Returns true if they match.
func (pg *passwordGenerator) Compare(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
