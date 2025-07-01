package passwd

import "golang.org/x/crypto/bcrypt"

type PasswordGenerator interface {
	Generate(string) (string, error)
	Compare(hash, password string) bool
}

type passwordGenerator struct {
	cost int
}

func NewPasswordGenerator(cost int) PasswordGenerator {
	return &passwordGenerator{cost: cost}
}

func (pg *passwordGenerator) Generate(password string) (string, error) {
	bts, err := bcrypt.GenerateFromPassword([]byte(password), pg.cost)
	if err != nil {
		return "", err
	}

	return string(bts), nil
}

func (pg *passwordGenerator) Compare(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
