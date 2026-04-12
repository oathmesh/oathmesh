package core

type ErrorCode string

const (
	ErrSignatureInvalid    ErrorCode = "signature_invalid"
	ErrIssuerUntrusted     ErrorCode = "issuer_untrusted"
	ErrTokenExpired        ErrorCode = "token_expired"
	ErrAudienceMismatch    ErrorCode = "audience_mismatch"
	ErrAlgorithmNotAllowed ErrorCode = "algorithm_not_allowed"
	ErrClaimMissing        ErrorCode = "claim_missing"
	ErrReplayDetected      ErrorCode = "replay_detected"
	ErrPolicyDenied        ErrorCode = "policy_denied"
	ErrBindingMismatch     ErrorCode = "binding_mismatch"
)

type OathMeshError struct {
	Code    ErrorCode `json:"error"`
	Message string    `json:"message"`
	Fix     string    `json:"fix"`
	ReqID   string    `json:"request_id,omitempty"`
}

func (e *OathMeshError) Error() string {
	return string(e.Code) + ": " + e.Message
}

func NewOathMeshError(code ErrorCode, message, fix string) *OathMeshError {
	return &OathMeshError{
		Code:    code,
		Message: message,
		Fix:     fix,
	}
}
