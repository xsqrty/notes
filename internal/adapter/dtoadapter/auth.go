package dtoadapter

import (
	"github.com/xsqrty/notes/internal/domain/auth"
	"github.com/xsqrty/notes/internal/domain/user"
	"github.com/xsqrty/notes/internal/dto"
)

// LoginRequestDtoToEntity converts a LoginRequest DTO to an auth.Login entity.
func LoginRequestDtoToEntity(request *dto.LoginRequest) *auth.Login {
	return &auth.Login{
		Email:    request.Email,
		Password: request.Password,
	}
}

// SignUpRequestDtoToEntity converts a SignUpRequest DTO to an auth.SignUp entity.
func SignUpRequestDtoToEntity(request *dto.SignUpRequest) *auth.SignUp {
	return &auth.SignUp{
		Email:    request.Email,
		Name:     request.Name,
		Password: request.Password,
	}
}

// TokensToResponseDto converts auth.Tokens to a dto.TokenResponse by mapping fields and calling UserToResponseDto.
func TokensToResponseDto(tokens *auth.Tokens) *dto.TokenResponse {
	return &dto.TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		User:         UserToResponseDto(tokens.User),
	}
}

// UserToResponseDto converts a user.User struct to a dto.UserResponse struct for external API responses.
func UserToResponseDto(user *user.User) *dto.UserResponse {
	return &dto.UserResponse{
		ID:    user.ID,
		Email: user.Email,
		Name:  user.Name,
	}
}
