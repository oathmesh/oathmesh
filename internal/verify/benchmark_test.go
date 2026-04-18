package verify

import (
	"context"
	"testing"
)

// BenchmarkVerify measures the pure CPU cost of verifying a token (14-step pipeline),
// specifically excluding ReplayCache (which would block the same token from being verified twice
// unless we pre-minted b.N tokens, which skews memory profiles).
//
// This benchmark exercises:
// 1. JWT parsing
// 2. Ed25519 strict signature verification
// 3. Claims validation
// 4. Policy evaluation (glob matching)
func BenchmarkVerify(b *testing.B) {
	privateKey, publicKey := generateTestKeys(b)
	cfg := testConfig(publicKey)

	// Omit replay cache so we can verify the exact same token in a tight loop
	cfg.ReplayCache = nil

	token := mintTestToken(b, privateKey, nil)
	ctx := context.Background()

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := Verify(ctx, token, cfg)
		if err != nil {
			b.Fatalf("verify failed: %v", err)
		}
	}
}
