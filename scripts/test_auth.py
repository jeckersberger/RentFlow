#!/usr/bin/env python3
"""Functional + security test suite for auth-service."""

import urllib.request
import json
import sys

BASE = "http://localhost:8001"


def req(method, path, body=None, headers=None):
    hdrs = {"Content-Type": "application/json"}
    if headers:
        hdrs.update(headers)
    data = json.dumps(body).encode() if body else None
    r = urllib.request.Request(BASE + path, data=data, headers=hdrs, method=method)
    try:
        resp = urllib.request.urlopen(r)
        raw = resp.read()
        return resp.status, json.loads(raw) if raw else {}
    except urllib.error.HTTPError as e:
        raw = e.read()
        try:
            return e.code, json.loads(raw) if raw else {}
        except (json.JSONDecodeError, ValueError):
            return e.code, {"error": raw.decode("utf-8", errors="replace")}


results = []

# 1. Health check
code, data = req("GET", "/health")
results.append(("Health", code == 200, code))

# 2. Setup init with empty token (should fail - Vuln 2 fix)
code, data = req("POST", "/api/v1/setup/init", {"token": ""})
results.append(("SECURITY: Setup empty token rejected", code != 200, code))

# 3. Setup init with valid token
code, data = req("POST", "/api/v1/setup/init", {"token": "test-setup-token-2026"})
results.append(("Setup init", code == 200, code))

# 4. Setup complete
code, data = req("POST", "/api/v1/setup/complete", {
    "token": "test-setup-token-2026",
    "company": {"name": "TestFirma", "slug": "testfirma", "email": "info@test.de", "phone": "+49123"},
    "admin": {"email": "admin@test.de", "password": "SecurePass123!", "first_name": "Max", "last_name": "Muster"},
    "industry": "event_technology",
})
results.append(("Setup complete", code in (200, 201), code))

# 5. Login
code, data = req("POST", "/api/v1/auth/login", {
    "email": "admin@test.de",
    "password": "SecurePass123!",
    "tenant_slug": "testfirma",
})
results.append(("Login", code == 200, code))
if code != 200:
    print("  LOGIN FAILED - cannot continue")
    sys.exit(1)

tokens = data.get("data", {}).get("tokens", {})
access = tokens.get("access_token", "")
refresh = tokens.get("refresh_token", "")
user_id = data.get("data", {}).get("user", {}).get("id", "")
auth_hdr = {"Authorization": "Bearer " + access}

# 6. Me endpoint
code, data = req("GET", "/api/v1/auth/me", headers=auth_hdr)
results.append(("Me", code == 200, code))

# 7. Update own profile
code, data = req("PUT", "/api/v1/auth/profile", {"phone": "+49999"}, headers=auth_hdr)
results.append(("Update profile", code == 200, code))

# 8. Create viewer user
code, data = req("POST", "/api/v1/users", {
    "email": "viewer@test.de",
    "password": "ViewerPass123!",
    "first_name": "View",
    "last_name": "Er",
    "role": "viewer",
}, headers=auth_hdr)
results.append(("Create user", code in (200, 201), code))
viewer_id = data.get("data", {}).get("id", "") if code in (200, 201) else ""

# 9. List users
code, data = req("GET", "/api/v1/users", headers=auth_hdr)
results.append(("List users", code == 200, code))

# 10. Get user by ID
if viewer_id:
    code, data = req("GET", "/api/v1/users/" + viewer_id, headers=auth_hdr)
    results.append(("Get user", code == 200, code))

# 11. Admin can update user
if viewer_id:
    code, data = req("PUT", "/api/v1/users/" + viewer_id, {"first_name": "Updated"}, headers=auth_hdr)
    results.append(("Admin update user", code == 200, code))

# 12. Update role
if viewer_id:
    code, data = req("PUT", "/api/v1/users/" + viewer_id + "/role", {"role": "user"}, headers=auth_hdr)
    results.append(("Update role", code in (200, 204), code))

# 13. Login as viewer - test Vuln 3 fix (role check)
code, data = req("POST", "/api/v1/auth/login", {
    "email": "viewer@test.de",
    "password": "ViewerPass123!",
    "tenant_slug": "testfirma",
})
if code == 200:
    viewer_access = data.get("data", {}).get("tokens", {}).get("access_token", "")
    viewer_hdr = {"Authorization": "Bearer " + viewer_access}

    # Viewer cannot update another user (requires admin)
    code2, _ = req("PUT", "/api/v1/users/" + user_id, {"first_name": "Hacked"}, headers=viewer_hdr)
    results.append(("SECURITY: Viewer cant update user", code2 == 403, code2))

    # Viewer cannot list users (requires admin/manager)
    code3, _ = req("GET", "/api/v1/users", headers=viewer_hdr)
    results.append(("SECURITY: Viewer cant list users", code3 == 403, code3))

    # Viewer cannot change roles
    code4, _ = req("PUT", "/api/v1/users/" + user_id + "/role", {"role": "admin"}, headers=viewer_hdr)
    results.append(("SECURITY: Viewer cant change role", code4 == 403, code4))

    # Viewer cannot delete users
    code5, _ = req("DELETE", "/api/v1/users/" + user_id, headers=viewer_hdr)
    results.append(("SECURITY: Viewer cant delete user", code5 == 403, code5))
else:
    results.append(("Viewer login", False, code))

# 14. Change password
code, data = req("PUT", "/api/v1/auth/password", {
    "old_password": "SecurePass123!",
    "new_password": "NewSecure456!",
}, headers=auth_hdr)
results.append(("Change password", code in (200, 204), code))

# 15. Old refresh token should be invalid after password change (sessions invalidated)
code, data = req("POST", "/api/v1/auth/refresh", {"refresh_token": refresh})
results.append(("SECURITY: Old refresh token invalid after pw change", code == 401, code))

# 16. Re-login after password change, deactivate viewer, logout
code, data = req("POST", "/api/v1/auth/login", {
    "email": "admin@test.de",
    "password": "NewSecure456!",
    "tenant_slug": "testfirma",
})
if code == 200:
    fresh_access = data.get("data", {}).get("tokens", {}).get("access_token", "")
    fresh_refresh = data.get("data", {}).get("tokens", {}).get("refresh_token", "")
    fresh_hdr = {"Authorization": "Bearer " + fresh_access}

    if viewer_id:
        code, data = req("DELETE", "/api/v1/users/" + viewer_id, headers=fresh_hdr)
        results.append(("Deactivate user", code in (200, 204), code))

    code, data = req("POST", "/api/v1/auth/logout", {"refresh_token": fresh_refresh}, headers=fresh_hdr)
    results.append(("Logout", code in (200, 204), code))

# 17. Setup init again should fail (already complete)
code, data = req("POST", "/api/v1/setup/init", {"token": "test-setup-token-2026"})
results.append(("SECURITY: Setup rejected after complete", code != 200, code))

# Print results
print("=" * 60)
print("AUTH SERVICE - FUNCTIONAL + SECURITY TESTS")
print("=" * 60)
passed = failed = 0
for name, ok, status_code in results:
    s = "PASS" if ok else "FAIL"
    if ok:
        passed += 1
    else:
        failed += 1
    print(f"  [{s}] {name} (HTTP {status_code})")
print("=" * 60)
print(f"  {passed} passed, {failed} failed out of {len(results)} tests")
if failed == 0:
    print("  ALL TESTS PASSED")
else:
    print("  SOME TESTS FAILED")
    sys.exit(1)
