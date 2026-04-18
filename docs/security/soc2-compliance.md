# SOC2 Compliance & Security Architecture

<p align="center">
  <img src="../../assets/logo.png" width="80" alt="OathMesh Logo">
</p>

This document natively maps OathMesh’s architectural components to the **AICPA Trust Services Criteria (SOC 2)**. It acts as an explicitly generated security matrix strictly designed for formal penetration testing scopes and third-party attestation audits.

---

## 1. Security (CC Series)

### 1.1 Access Control Boundaries (CC6.1 - CC6.3)
OathMesh aggressively enforces zero-trust identity dynamically restricting access at exactly the 14-step boundary check.

- **Pre-Shared Key Secrecy**: The Control Plane / Admin API (e.g. `POST /v1/token`, `POST /v1/admin/revoke`) is strictly gated behind `OATHMESH_MINT_SECRET` securely preventing untrusted issuers entirely dynamically mapping internal network scopes.
- **Fail-Closed Networking**: The `internal/gateway` operates as a reverse proxy cleanly completely disconnecting untrusted client IPs from the upstream network natively. Validation completely isolates unverified packets and returns HTTP `401 Unauthorized`.
- **Dynamic OIDC Auto-Sign**: Ephemeral identity contexts natively exchanged by AWS/GitHub/GitLab securely enforce automated CI pipelines dynamically.

### 1.2 Cryptographic Standard (CC6.7)
Data perfectly transitions exactly bounded entirely securely across environments natively:

| Control | Implementation Details |
|---------|-----------------------|
| Algorithm | Asymmetric `Ed25519` (EdDSA) |
| Signatures| Hardware-Backed AWS KMS (via `MessageType: RAW`) using natively strictly configured hardware, completely avoiding node-local persistence vulnerabilities. |
| Header Inject | `X-Oathmesh-*` bounded headers explicitly mapped verifying contexts |

---

## 2. Integrity & Isolation (PI Series)

### 2.1 Multi-Tenant Separation (PI1.4)
Strict bounding formally asserts environment contexts globally.
- OathMesh natively pushes the `tenant` schema exactly mapped into the token representations preventing IDOR vulnerabilities locally. 

### 2.2 Replay Protection Cache (PI1.3)
All assertions explicitly contain unique `JTI` tokens inherently rejecting identical strings dynamically across the Redis cluster memory space spanning a `max 300 seconds TTL` TTL securely. 

---

## Penetration Testing Scopes

### Permitted Targets
- **Gateway Reverse Proxy:** Injection vectors scaling across the `Authorization` header schemas natively mapping. 
- **OIDC Discovery Matrix:** Enumeration scopes extracting explicitly defined JWKS configurations locally. 
- **Time-Based Attacks:** Clock skew manipulation accurately strictly measuring exact parsing boundaries (`time.Now()` overrides).

### Excluded Scopes
- AWS KMS underlying hardware infrastructure configurations logically.
- Core language vulnerabilities mapping directly tracking Go 1.25.0 vulnerabilities aggressively handled explicitly.

### Disclosure 
External researchers verifying scopes exactly mappings to `security@oathmesh.dev` dynamically securely tracking reports aggressively responding under 48 hours natively.
