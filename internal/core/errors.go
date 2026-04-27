package core

type ErrorCode string

const (
	ErrSignatureInvalid    ErrorCode = "signature_invalid"
	ErrIssuerUntrusted     ErrorCode = "issuer_untrusted"
	ErrTokenExpired        ErrorCode = "token_expired"
	ErrAudienceMismatch    ErrorCode = "audience_mismatch"
	ErrAlgorithmNotAllowed ErrorCode = "algorithm_not_allowed"
	ErrTokenMalformed      ErrorCode = "token_malformed"
	ErrClaimMissing        ErrorCode = "claim_missing"
	ErrReplayDetected      ErrorCode = "replay_detected"
	ErrPolicyDenied        ErrorCode = "policy_denied"
	ErrBindingMismatch     ErrorCode = "binding_mismatch"
	ErrBindingRequired     ErrorCode = "binding_required"
	ErrSubjectRevoked      ErrorCode = "subject_revoked"
	ErrVerificationFailed  ErrorCode = "verification_failed"
)

type OathMeshError struct {
	Code    ErrorCode `json:"error"`
	Message string    `json:"message"`
	Fix     string    `json:"fix"`
	ReqID   string    `json:"request_id,omitempty"`
	Step    int       `json:"step,omitempty"`
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

// NewOathMeshErrorAt creates an OathMeshError annotated with the verification
// step number that triggered the failure. This enables operators to immediately
// identify where in the 14-step pipeline a token was rejected.
func NewOathMeshErrorAt(step int, code ErrorCode, message, fix string) *OathMeshError {
	return &OathMeshError{
		Code:    code,
		Message: message,
		Fix:     fix,
		Step:    step,
	}
}
