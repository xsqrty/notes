package dtoadapter

import (
	"github.com/xsqrty/notes/internal/domain/auth"
	"github.com/xsqrty/notes/internal/domain/user"
	"github.com/xsqrty/notes/internal/dto"
)

func LoginRequestDtoToEntity(request *dto.LoginRequest) *auth.Login {
	return &auth.Login{
		Email:    request.Email,
		Password: request.Password,
	}
}

func SignUpRequestDtoToEntity(request *dto.SignUpRequest) *auth.SignUp {
	return &auth.SignUp{
		Email:    request.Email,
		Name:     request.Name,
		Password: request.Password,
	}
}

func TokensToResponseDto(tokens *auth.Tokens) *dto.TokenResponse {
	return &dto.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		User:         UserToResponseDto(tokens.User),
	}
}

func UserToResponseDto(user *user.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	}
}
