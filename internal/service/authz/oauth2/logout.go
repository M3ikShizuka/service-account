package oauth2

import "golang.org/x/net/context"

func (h *HandlerOAuth2) RejectLogoutRequest(context context.Context, challenge string) error {
	request := h.hydra.AdminApi.RejectLogoutRequest(context)
	request = request.LogoutChallenge(challenge)
	_, err := request.Execute()
	if err != nil {
		// Error request to hydra OAuth admin API.
		return err
	}

	return nil
}

func (h *HandlerOAuth2) AcceptLogoutRequest(context context.Context, challenge string) (string, error) {
	requestAcceptLogout := h.hydra.AdminApi.AcceptLogoutRequest(context)
	requestAcceptLogout = requestAcceptLogout.LogoutChallenge(challenge)
	completedRequest, _, errAcceptLogout := requestAcceptLogout.Execute()
	if errAcceptLogout != nil {
		// Error request to hydra OAuth admin API.
		return "", nil
	}

	return completedRequest.RedirectTo, nil
}
