#!/usr/bin/env python3
"""Functional test suite for scanner-service."""

import urllib.request
import json
import sys

BASE = "http://localhost:8004"
AUTH_BASE = "http://localhost:8001"


def req(method, path, body=None, headers=None, base=None):
    hdrs = {"Content-Type": "application/json"}
    if headers:
        hdrs.update(headers)
    data = json.dumps(body).encode() if body else None
    url = (base or BASE) + path
    r = urllib.request.Request(url, data=data, headers=hdrs, method=method)
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

# 0. Get auth token
code, data = req("POST", "/api/v1/auth/login", {
    "email": "j.eckersberger@je-soundulight.de",
    "password": "RentFlow2026!",
    "tenant_slug": "je-soundulight",
}, base=AUTH_BASE)
if code != 200:
    print(f"AUTH FAILED (HTTP {code}) — trying setup first...")
    # Try setup
    req("POST", "/api/v1/setup/init", {"token": "test-setup-token-2026"}, base=AUTH_BASE)
    req("POST", "/api/v1/setup/complete", {
        "token": "test-setup-token-2026",
        "company": {"name": "JE Sound & Light", "slug": "je-soundulight", "email": "j.eckersberger@je-soundulight.de", "phone": "+49123"},
        "admin": {"email": "j.eckersberger@je-soundulight.de", "password": "RentFlow2026!", "first_name": "Janis", "last_name": "Eckersberger"},
        "industry": "event_technology",
    }, base=AUTH_BASE)
    code, data = req("POST", "/api/v1/auth/login", {
        "email": "j.eckersberger@je-soundulight.de",
        "password": "RentFlow2026!",
        "tenant_slug": "je-soundulight",
    }, base=AUTH_BASE)
    if code != 200:
        print(f"  LOGIN FAILED - cannot continue (HTTP {code})")
        sys.exit(1)

access = data.get("data", {}).get("tokens", {}).get("access_token", "")
auth_hdr = {"Authorization": "Bearer " + access}

# 1. Health check
code, data = req("GET", "/health")
results.append(("Health", code == 200, code))

# 2. Scan without auth -> 401
code, data = req("POST", "/api/v1/scanner/scan", {"barcode": "EQ-001"})
results.append(("SECURITY: Scan without auth rejected", code == 401, code))

# 3. Scan with barcode
code, data = req("POST", "/api/v1/scanner/scan", {
    "device_id": "DEV-TEST-001",
    "barcode": "EQ-001",
}, headers=auth_hdr)
results.append(("Scan barcode", code == 201, code))
scan_event_id = data.get("data", {}).get("id", "") if code == 201 else ""

# 4. Scan with RFID
code, data = req("POST", "/api/v1/scanner/scan", {
    "device_id": "DEV-TEST-001",
    "rfid_tag": "RFID-TAG-001",
}, headers=auth_hdr)
results.append(("Scan RFID", code == 201, code))

# 5. Scan without identifier -> 400
code, data = req("POST", "/api/v1/scanner/scan", {
    "device_id": "DEV-TEST-001",
}, headers=auth_hdr)
results.append(("Scan without identifier rejected", code == 400, code))

# 6. Register device
code, data = req("POST", "/api/v1/scanner/devices/register", {
    "device_id": "CF-H906-TEST-001",
    "device_name": "Test Scanner",
    "device_type": "cf-h906",
}, headers=auth_hdr)
results.append(("Register device", code == 201, code))
device_db_id = data.get("data", {}).get("id", "") if code == 201 else ""

# 7. List devices
code, data = req("GET", "/api/v1/scanner/devices", headers=auth_hdr)
results.append(("List devices", code == 200, code))
devices = data.get("data", [])
results.append(("Devices not empty", isinstance(devices, list) and len(devices) > 0, len(devices) if isinstance(devices, list) else 0))

# 8. Trigger ring
if device_db_id:
    code, data = req("POST", f"/api/v1/scanner/devices/{device_db_id}/ring", headers=auth_hdr)
    results.append(("Trigger ring", code == 204, code))

# 9. Check ring
code, data = req("GET", "/api/v1/scanner/devices/ring?device_id=CF-H906-TEST-001", headers=auth_hdr)
results.append(("Check ring", code == 200, code))
if code == 200:
    ring_requested = data.get("data", {}).get("ring_requested", False)
    results.append(("Ring is requested", ring_requested is True, ring_requested))

# 10. Ack ring
code, data = req("POST", "/api/v1/scanner/devices/ring-ack", {
    "device_id": "CF-H906-TEST-001",
}, headers=auth_hdr)
results.append(("Ack ring", code == 204, code))

# 11. Checkout (bulk)
import uuid as uuid_mod
equip_id = str(uuid_mod.uuid4())
project_id = str(uuid_mod.uuid4())
code, data = req("POST", "/api/v1/scanner/checkout", {
    "device_id": "DEV-TEST-001",
    "project_id": project_id,
    "items": [
        {"equipment_id": equip_id, "barcode": "EQ-CHECKOUT-001"},
    ],
}, headers=auth_hdr)
results.append(("Checkout", code == 201, code))

# 12. Checkin (bulk with rating)
code, data = req("POST", "/api/v1/scanner/checkin", {
    "device_id": "DEV-TEST-001",
    "items": [
        {"equipment_id": equip_id, "barcode": "EQ-CHECKOUT-001", "condition_rating": 4, "condition_notes": "Alles OK"},
    ],
}, headers=auth_hdr)
results.append(("Checkin with rating", code == 201, code))

# 13. Adhoc booking
code, data = req("POST", "/api/v1/scanner/adhoc-booking", {
    "device_id": "DEV-TEST-001",
    "barcode": "EQ-ADHOC-001",
}, headers=auth_hdr)
results.append(("Adhoc booking", code == 201, code))

# 14. Bulk sync
import datetime
ts = datetime.datetime.now(datetime.timezone.utc).isoformat()
code, data = req("POST", "/api/v1/scanner/bulk", {
    "events": [
        {"device_id": "DEV-TEST-001", "barcode": "EQ-BULK-001", "action": "scan", "timestamp": ts},
        {"device_id": "DEV-TEST-001", "rfid_tag": "RFID-BULK-001", "action": "scan", "timestamp": "2025-01-01T00:00:00Z"},
    ],
}, headers=auth_hdr)
results.append(("Bulk sync", code == 200, code))
if code == 200:
    inserted = data.get("data", {}).get("inserted", 0)
    results.append(("Bulk sync inserted 2", inserted == 2, inserted))

# 15. Bulk sync again (dedup)
code, data = req("POST", "/api/v1/scanner/bulk", {
    "events": [
        {"device_id": "DEV-TEST-001", "barcode": "EQ-BULK-001", "action": "scan", "timestamp": ts},
    ],
}, headers=auth_hdr)
results.append(("Bulk sync dedup", code == 200, code))
if code == 200:
    skipped = data.get("data", {}).get("skipped", 0)
    results.append(("Bulk sync dedup skipped", skipped >= 1, skipped))

# 16. List events
code, data = req("GET", "/api/v1/scanner/events", headers=auth_hdr)
results.append(("List events", code == 200, code))

# 17. List events with filter
code, data = req("GET", "/api/v1/scanner/events?action=scan", headers=auth_hdr)
results.append(("List events filtered", code == 200, code))

# Print results
print("=" * 60)
print("SCANNER SERVICE - FUNCTIONAL TESTS")
print("=" * 60)
passed = failed = 0
for name, ok, status_code in results:
    s = "PASS" if ok else "FAIL"
    if ok:
        passed += 1
    else:
        failed += 1
    print(f"  [{s}] {name} ({status_code})")
print("=" * 60)
print(f"  {passed} passed, {failed} failed out of {len(results)} tests")
if failed == 0:
    print("  ALL TESTS PASSED")
else:
    print("  SOME TESTS FAILED")
    sys.exit(1)
