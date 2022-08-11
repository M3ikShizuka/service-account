package oauth2

import (
	"golang.org/x/net/context"
	"service-account/internal/domain"
)

func (h *HandlerOAuth2) IntrospectOAuth2Token(context context.Context, accessToken string) (*domain.OA2TokenIntrospection, error) {
	requestIntrospectToken := h.hydra.AdminApi.IntrospectOAuth2Token(context)
	requestIntrospectToken = requestIntrospectToken.Token(accessToken)
	tokenIntrospection, _, err := requestIntrospectToken.Execute()
	if err != nil {
		// Error request to hydra OAuth admin API.
		return nil, err
	}

	return &domain.OA2TokenIntrospection{
		Active: tokenIntrospection.GetActive(),
	}, nil
}
