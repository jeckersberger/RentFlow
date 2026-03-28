#!/usr/bin/env python3
"""Functional test suite for warehouse-service."""

import urllib.request
import json
import sys

BASE = "http://localhost:8005"
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

# 2. Create warehouse without auth -> 401
code, data = req("POST", "/api/v1/warehouses/", {"name": "Test"})
results.append(("SECURITY: Create warehouse without auth rejected", code == 401, code))

# 3. Create warehouse
code, data = req("POST", "/api/v1/warehouses/", {
    "name": "Hauptlager",
    "code": "HL-001",
    "address": "Musterstr. 1, 12345 Berlin",
    "capacity_description": "500 qm Hallenflaeche",
}, headers=auth_hdr)
results.append(("Create warehouse", code == 201, code))
warehouse_id = data.get("data", {}).get("id", "") if code == 201 else ""

# 4. Create warehouse without name -> 400
code, data = req("POST", "/api/v1/warehouses/", {"code": "X"}, headers=auth_hdr)
results.append(("Create warehouse without name rejected", code == 400, code))

# 5. List warehouses
code, data = req("GET", "/api/v1/warehouses/", headers=auth_hdr)
results.append(("List warehouses", code == 200, code))
warehouses = data.get("data", [])
results.append(("Warehouses not empty", isinstance(warehouses, list) and len(warehouses) > 0, len(warehouses) if isinstance(warehouses, list) else 0))

# 6. Get warehouse by ID
if warehouse_id:
    code, data = req("GET", f"/api/v1/warehouses/{warehouse_id}", headers=auth_hdr)
    results.append(("Get warehouse by ID", code == 200, code))

# 7. Create zone
code, data = req("POST", "/api/v1/zones/", {
    "warehouse_id": warehouse_id,
    "name": "Buehnen-Zone",
    "code": "BZ-001",
    "climate_controlled": False,
    "notes": "Fuer Buehnenequipment",
}, headers=auth_hdr)
results.append(("Create zone", code == 201, code))
zone_id = data.get("data", {}).get("id", "") if code == 201 else ""

# 8. Create zone without name -> 400
code, data = req("POST", "/api/v1/zones/", {
    "warehouse_id": warehouse_id,
}, headers=auth_hdr)
results.append(("Create zone without name rejected", code == 400, code))

# 9. List zones
code, data = req("GET", "/api/v1/zones/", headers=auth_hdr)
results.append(("List zones", code == 200, code))

# 10. Update zone
if zone_id:
    code, data = req("PUT", f"/api/v1/zones/{zone_id}", {
        "name": "Buehnen-Zone (aktualisiert)",
        "climate_controlled": True,
    }, headers=auth_hdr)
    results.append(("Update zone", code == 200, code))

# 11. Create rack
code, data = req("POST", "/api/v1/racks/", {
    "zone_id": zone_id,
    "name": "Regal A",
    "code": "RA-001",
    "levels": 5,
    "bays_per_level": 8,
}, headers=auth_hdr)
results.append(("Create rack", code == 201, code))
rack_id = data.get("data", {}).get("id", "") if code == 201 else ""

# 12. List racks
code, data = req("GET", "/api/v1/racks/", headers=auth_hdr)
results.append(("List racks", code == 200, code))

# 13. Create stock location
code, data = req("POST", "/api/v1/locations/", {
    "rack_id": rack_id,
    "zone_id": zone_id,
    "code": "RA-001-L3-B5",
    "barcode": "LOC-RA001L3B5",
    "level": 3,
    "bay": 5,
}, headers=auth_hdr)
results.append(("Create location", code == 201, code))
location_id = data.get("data", {}).get("id", "") if code == 201 else ""

# 14. Duplicate location code -> 409
code, data = req("POST", "/api/v1/locations/", {
    "rack_id": rack_id,
    "zone_id": zone_id,
    "code": "RA-001-L3-B5",
}, headers=auth_hdr)
results.append(("Duplicate location code rejected", code == 409, code))

# 15. List locations
code, data = req("GET", "/api/v1/locations/", headers=auth_hdr)
results.append(("List locations", code == 200, code))

# 16. Get location by ID
if location_id:
    code, data = req("GET", f"/api/v1/locations/{location_id}", headers=auth_hdr)
    results.append(("Get location by ID", code == 200, code))

# 17. Create movement
import uuid as uuid_mod
equip_id = str(uuid_mod.uuid4())
code, data = req("POST", "/api/v1/movements/", {
    "equipment_id": equip_id,
    "to_location_id": location_id,
    "quantity": 1,
    "reason": "Einlagerung",
    "notes": "Ersteinlagerung nach Lieferung",
}, headers=auth_hdr)
results.append(("Create movement", code == 201, code))

# 18. List movements
code, data = req("GET", "/api/v1/movements/", headers=auth_hdr)
results.append(("List movements", code == 200, code))

# 19. Create inventory check
code, data = req("POST", "/api/v1/inventory-checks/", {
    "zone_id": zone_id,
    "expected_count": 10,
    "notes": "Quartalsinventur Q1",
}, headers=auth_hdr)
results.append(("Create inventory check", code == 201, code))
check_id = data.get("data", {}).get("id", "") if code == 201 else ""

# 20. List inventory checks
code, data = req("GET", "/api/v1/inventory-checks/", headers=auth_hdr)
results.append(("List inventory checks", code == 200, code))

# 21. Scan item into inventory check
if check_id:
    code, data = req("POST", f"/api/v1/inventory-checks/{check_id}/scan", {
        "equipment_id": equip_id,
        "notes": "Gefunden an erwarteter Position",
    }, headers=auth_hdr)
    results.append(("Scan inventory item", code in (200, 201), code))

# 22. Complete inventory check
if check_id:
    code, data = req("PATCH", f"/api/v1/inventory-checks/{check_id}/complete", headers=auth_hdr)
    results.append(("Complete inventory check", code in (200, 204), code))

# 23. Complete again -> 409
if check_id:
    code, data = req("PATCH", f"/api/v1/inventory-checks/{check_id}/complete", headers=auth_hdr)
    results.append(("Complete again rejected", code == 409, code))

# 24. Get inventory check result
if check_id:
    code, data = req("GET", f"/api/v1/inventory-checks/{check_id}/result", headers=auth_hdr)
    results.append(("Get inventory check result", code == 200, code))

# Print results
print("=" * 60)
print("WAREHOUSE SERVICE - FUNCTIONAL TESTS")
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
