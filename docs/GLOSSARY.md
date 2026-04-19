# OathMesh Glossary

## A

**Action (`act`)**  
The declared operation a caller wants to perform (for example `invoice.read`), used by policy evaluation.

**Audience (`aud`)**  
The intended receiver of a token. Tokens must be rejected if the audience does not match the target service.

**Audit Event**  
A structured log record describing a security-relevant action or decision, such as token verification allow/deny.

## C

**Caller**  
The workload or service making a machine-to-machine request with an OathMesh token.

**Claim**  
A key-value field inside a token payload, such as `sub`, `aud`, `act`, `iat`, `exp`, or `jti`.

**Compliance Evidence**  
Artifacts proving controls are operating (logs, approvals, runbooks, reports, and review records).

## D

**Default Deny**  
Authorization posture where requests are denied unless explicitly permitted by policy.

## E

**Ed25519**  
The asymmetric signature algorithm used by OathMesh for signing and verification.

**Expiry (`exp`)**  
The claim that defines when a token is no longer valid.

## G

**Gateway Mode**  
Deployment pattern where OathMesh verification happens at an ingress/proxy layer before requests reach services.

## I

**Issuer**  
The component that validates token requests and signs short-lived tokens.

**Issued At (`iat`)**  
The claim indicating when a token was created.

## J

**JWT (JSON Web Token)**  
A compact, signed token format used as the envelope for OathMesh assertions.

**JTI (`jti`)**  
A unique token identifier used to detect and block replay attempts.

## K

**Key Rotation**  
The controlled process of replacing signing keys and retiring old keys safely.

## M

**Machine Identity**  
Cryptographic identity used by software workloads to authenticate service-to-service calls.

## N

**NDJSON**  
Newline-delimited JSON format used for streaming structured audit events.

## O

**Oath Token**  
A short-lived signed token carrying caller identity, intended audience, and requested action.

## P

**Policy Engine**  
The subsystem that evaluates request context against allow/deny rules.

**Provenance**  
Metadata describing call origin (for example CI run, workload source, environment, or attested context).

## R

**Receiver**  
The service (or protected endpoint) that verifies tokens and enforces authorization decisions.

**Replay Attack**  
An attacker reuses a previously valid token to gain unauthorized access.

**Replay Cache**  
Storage used to remember used token IDs (`jti`) for the token validity window.

## S

**Scope**  
The bounded permission intent represented by token claims and policy conditions.

**Subject (`sub`)**  
The identity asserted by the token, usually a workload, service account, or machine principal.

## T

**TTL (Time To Live)**  
Maximum lifetime of a token before expiry; OathMesh tokens are intentionally short-lived.

**Trust Boundary**  
A point where security assumptions change and explicit verification is required.

## V

**Verification Pipeline**  
The ordered checks performed by the receiver before accepting a token (format, signature, claims, replay, policy).
