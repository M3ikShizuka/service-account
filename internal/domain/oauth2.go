package domain

// LoginRequest struct for LoginRequest
type OA2LoginRequest struct {
	// ID is the identifier (\"login challenge\") of the login request. It is used to identify the session.
	Challenge   string                `json:"challenge"`
	Client      OAuth2Client          `json:"client"`
	OidcContext *OpenIDConnectContext `json:"oidc_context,omitempty"`
	// RequestURL is the original OAuth 2.0 Authorization URL requested by the OAuth 2.0 client. It is the URL which initiates the OAuth 2.0 Authorization Code or OAuth 2.0 Implicit flow. This URL is typically not needed, but might come in handy if you want to deal with additional request parameters.
	RequestUrl                   string   `json:"request_url"`
	RequestedAccessTokenAudience []string `json:"requested_access_token_audience"`
	RequestedScope               []string `json:"requested_scope"`
	// SessionID is the login session ID. If the user-agent reuses a login session (via cookie / remember flag) this ID will remain the same. If the user-agent did not have an existing authentication session (e.g. remember is false) this will be a new random value. This value is used as the \"sid\" parameter in the ID Token and in OIDC Front-/Back- channel logout. It's value can generally be used to associate consecutive login requests by a certain user.
	SessionId *string `json:"session_id,omitempty"`
	// Skip, if true, implies that the client has requested the same scopes from the same user previously. If true, you can skip asking the user to grant the requested scopes, and simply forward the user to the redirect URL.  This feature allows you to update / set session information.
	Skip bool
	// Subject is the user ID of the end-user that authenticated. Now, that end user needs to grant or deny the scope requested by the OAuth 2.0 client. If this value is set and `skip` is true, you MUST include this subject type when accepting the login request, or the request will fail.
	Subject string `json:"subject"`
}
