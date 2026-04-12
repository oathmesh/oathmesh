import pytest
import sys
import os
sys.path.insert(0, os.path.abspath(os.path.join(os.path.dirname(__file__), '../src')))

from oathmesh.verify import verify_token, VerifierConfig
from oathmesh.errors import OathMeshError

def test_missing_authorization():
    config = VerifierConfig(audience="audience", trusted_issuers=["http://issuer.local"])
    with pytest.raises(OathMeshError) as exc_info:
        verify_token("", config)
    assert exc_info.value.code == "claim_missing:token"

def test_malformed_token():
    config = VerifierConfig(audience="audience", trusted_issuers=["http://issuer.local"])
    with pytest.raises(OathMeshError) as exc_info:
        verify_token("OathMesh badlyformed", config)
    assert exc_info.value.code == "verification_failed"

def test_wrong_authorization_form():
    config = VerifierConfig(audience="audience", trusted_issuers=["http://issuer.local"])
    with pytest.raises(OathMeshError) as exc_info:
        verify_token("Bearer something", config)
    assert exc_info.value.code == "claim_missing:token"
