#!/usr/bin/env python3
"""
CrateDesk Inventory Service — Functional & Security Tests
Tests all CRUD operations for Equipment, Categories, Equipment Types, and Flightcases.
"""

import json
import sys
import uuid
import requests

AUTH_URL = "http://localhost:8001"
INV_URL = "http://localhost:8002"
EMAIL = "j.eckersberger@je-soundulight.de"
PASSWORD = "RentFlow2026!"
TENANT_SLUG = "je-soundulight"

passed = 0
failed = 0
errors = []


def test(name, condition, detail=""):
    global passed, failed
    if condition:
        passed += 1
        print(f"  [PASS] {name}")
    else:
        failed += 1
        errors.append(name)
        print(f"  [FAIL] {name} — {detail}")


def login():
    r = requests.post(f"{AUTH_URL}/api/v1/auth/login", json={
        "email": EMAIL, "password": PASSWORD, "tenant_slug": TENANT_SLUG
    })
    data = r.json()
    token = data.get("data", {}).get("tokens", {}).get("access_token", "")
    return token


def headers(token):
    return {"Authorization": f"Bearer {token}", "Content-Type": "application/json"}


def main():
    print("=" * 60)
    print("CrateDesk Inventory Service — Test Suite")
    print("=" * 60)

    # --- Auth ---
    print("\n[1] Authentication")
    token = login()
    test("Login successful", token != "", "No token returned")
    h = headers(token)

    # --- Health ---
    print("\n[2] Health Checks")
    r = requests.get(f"{INV_URL}/health")
    test("Health endpoint", r.status_code == 200)
    r = requests.get(f"{INV_URL}/livez")
    test("Liveness endpoint", r.status_code == 200)

    # Unique suffix per test run to avoid duplicate conflicts.
    uid = uuid.uuid4().hex[:6]

    # --- Categories ---
    print("\n[3] Categories CRUD")
    r = requests.post(f"{INV_URL}/api/v1/categories", headers=h, json={
        "name": f"Audio-{uid}", "icon": "speaker", "color": "#FF5500", "sort_order": 1
    })
    test("Create category", r.status_code == 201, f"status={r.status_code} body={r.text[:200]}")
    cat_id = r.json().get("data", {}).get("id") if r.status_code == 201 else None

    r = requests.post(f"{INV_URL}/api/v1/categories", headers=h, json={
        "name": f"Licht-{uid}", "icon": "lightbulb", "color": "#00AAFF", "sort_order": 2
    })
    test("Create second category", r.status_code == 201, f"status={r.status_code}")
    cat2_id = r.json().get("data", {}).get("id") if r.status_code == 201 else None

    r = requests.get(f"{INV_URL}/api/v1/categories", headers=h)
    test("List categories", r.status_code == 200 and len(r.json().get("data", [])) >= 2,
         f"status={r.status_code} count={len(r.json().get('data', []))}")

    if cat_id:
        r = requests.get(f"{INV_URL}/api/v1/categories/{cat_id}", headers=h)
        test("Get category by ID", r.status_code == 200 and f"Audio-{uid}" in r.json()["data"]["name"])

        r = requests.put(f"{INV_URL}/api/v1/categories/{cat_id}", headers=h, json={
            "name": f"Audio-PA-{uid}"
        })
        test("Update category", r.status_code == 200 and f"Audio-PA-{uid}" in r.json()["data"]["name"],
             f"status={r.status_code} body={r.text[:200]}")

    # --- Equipment Types ---
    print("\n[4] Equipment Types CRUD")
    r = requests.post(f"{INV_URL}/api/v1/equipment-types", headers=h, json={
        "name": f"QSC K12.2-{uid}",
        "description": "12-inch powered speaker",
        "category_id": cat_id,
        "default_rental_price_day": 4500,
        "default_rental_price_week": 18000,
        "default_replacement_value": 95000
    })
    test("Create equipment type", r.status_code == 201, f"status={r.status_code} body={r.text[:200]}")
    type_id = r.json().get("data", {}).get("id") if r.status_code == 201 else None

    r = requests.get(f"{INV_URL}/api/v1/equipment-types", headers=h)
    test("List equipment types", r.status_code == 200)

    if type_id:
        r = requests.get(f"{INV_URL}/api/v1/equipment-types/{type_id}", headers=h)
        test("Get equipment type by ID", r.status_code == 200)

        r = requests.put(f"{INV_URL}/api/v1/equipment-types/{type_id}", headers=h, json={
            "name": "QSC K12.2 Active"
        })
        test("Update equipment type", r.status_code == 200,
             f"status={r.status_code} body={r.text[:200]}")

    # --- Bulk Create from Type ---
    print("\n[5] Bulk Create Equipment from Type")
    if type_id:
        r = requests.post(f"{INV_URL}/api/v1/equipment-types/{type_id}/create-items", headers=h, json={
            "count": 3,
            "name_prefix": f"QSC-{uid}",
            "start_number": 1
        })
        test("Bulk create 3 items", r.status_code == 201 and len(r.json().get("data", [])) == 3,
             f"status={r.status_code} body={r.text[:300]}")

    # --- Equipment CRUD ---
    print("\n[6] Equipment CRUD")
    barcode = f"SM58-{uid}"
    rfid_tag = f"E200-{uid}-1234-5678"
    r = requests.post(f"{INV_URL}/api/v1/equipment", headers=h, json={
        "name": f"Shure SM58-{uid}",
        "description": "Dynamisches Gesangsmikrofon",
        "category_id": cat_id,
        "barcode": barcode,
        "serial_number": f"SN-SM58-{uid}",
        "rental_price_day": 1500,
        "rental_price_week": 6000,
        "replacement_value": 12000,
        "quantity_total": 1,
        "manufacturer": "Shure",
        "model": "SM58"
    })
    test("Create equipment", r.status_code == 201, f"status={r.status_code} body={r.text[:300]}")
    eq_id = r.json().get("data", {}).get("id") if r.status_code == 201 else None

    r = requests.get(f"{INV_URL}/api/v1/equipment?page=1&per_page=50", headers=h)
    test("List equipment", r.status_code == 200, f"status={r.status_code}")

    if eq_id:
        r = requests.get(f"{INV_URL}/api/v1/equipment/{eq_id}", headers=h)
        test("Get equipment by ID", r.status_code == 200 and f"SM58-{uid}" in r.json()["data"]["name"])

        r = requests.put(f"{INV_URL}/api/v1/equipment/{eq_id}", headers=h, json={
            "name": f"Shure SM58 #{uid}"
        })
        test("Update equipment", r.status_code == 200,
             f"status={r.status_code} body={r.text[:200]}")

        # Status update
        r = requests.patch(f"{INV_URL}/api/v1/equipment/{eq_id}/status", headers=h, json={
            "status": "checked_out"
        })
        test("Update status to checked_out", r.status_code == 204, f"status={r.status_code} body={r.text[:200]}")

        # Condition update
        r = requests.patch(f"{INV_URL}/api/v1/equipment/{eq_id}/condition", headers=h, json={
            "condition": "damaged"
        })
        test("Update condition to damaged", r.status_code == 204,
             f"status={r.status_code} body={r.text[:200]}")

        # RFID assign
        r = requests.patch(f"{INV_URL}/api/v1/equipment/{eq_id}/rfid", headers=h, json={
            "rfid_tag": rfid_tag
        })
        test("Assign RFID tag", r.status_code == 204, f"status={r.status_code} body={r.text[:200]}")

        # History
        r = requests.get(f"{INV_URL}/api/v1/equipment/{eq_id}/history", headers=h)
        test("Get equipment history", r.status_code == 200, f"status={r.status_code}")
        if r.status_code == 200:
            history_count = len(r.json().get("data", []))
            test("History has entries", history_count > 0, f"count={history_count}")

    # --- Barcode & RFID Lookup ---
    print("\n[7] Barcode & RFID Lookup")
    if eq_id:
        r = requests.get(f"{INV_URL}/api/v1/equipment/lookup/barcode/{barcode}", headers=h)
        test("Lookup by barcode", r.status_code == 200, f"status={r.status_code}")

        r = requests.get(f"{INV_URL}/api/v1/equipment/lookup/rfid/{rfid_tag}", headers=h)
        test("Lookup by RFID", r.status_code == 200, f"status={r.status_code}")

    # --- Search ---
    print("\n[8] Full-Text Search")
    r = requests.get(f"{INV_URL}/api/v1/equipment/search?q=Shure", headers=h)
    test("Search for 'Shure'", r.status_code == 200, f"status={r.status_code}")

    # --- Flightcases ---
    print("\n[9] Flightcases CRUD")
    r = requests.post(f"{INV_URL}/api/v1/flightcases", headers=h, json={
        "name": f"Audio Rack-{uid}",
        "barcode": f"FC-{uid}",
        "description": "Standard Audio Rack fuer PA",
        "weight_grams": 25000
    })
    test("Create flightcase", r.status_code == 201, f"status={r.status_code} body={r.text[:200]}")
    fc_id = r.json().get("data", {}).get("id") if r.status_code == 201 else None

    r = requests.get(f"{INV_URL}/api/v1/flightcases?page=1&per_page=50", headers=h)
    test("List flightcases", r.status_code == 200, f"status={r.status_code}")

    if fc_id:
        r = requests.get(f"{INV_URL}/api/v1/flightcases/{fc_id}", headers=h)
        test("Get flightcase by ID", r.status_code == 200)

        r = requests.put(f"{INV_URL}/api/v1/flightcases/{fc_id}", headers=h, json={
            "name": "Audio Rack 1 (Updated)"
        })
        test("Update flightcase", r.status_code == 200, f"status={r.status_code} body={r.text[:200]}")

        # Add item to flightcase
        if eq_id:
            r = requests.post(f"{INV_URL}/api/v1/flightcases/{fc_id}/items", headers=h, json={
                "equipment_id": eq_id,
                "quantity": 1,
                "sort_order": 1
            })
            test("Add item to flightcase", r.status_code == 201,
                 f"status={r.status_code} body={r.text[:200]}")
            item_id = r.json().get("data", {}).get("id") if r.status_code == 201 else None

            # Get items
            r = requests.get(f"{INV_URL}/api/v1/flightcases/{fc_id}/items", headers=h)
            test("Get flightcase items", r.status_code == 200 and len(r.json().get("data", [])) >= 1,
                 f"status={r.status_code}")

            # Remove item
            if item_id:
                r = requests.delete(f"{INV_URL}/api/v1/flightcases/{fc_id}/items/{item_id}", headers=h)
                test("Remove item from flightcase", r.status_code == 204,
                     f"status={r.status_code}")

    # --- Security Tests ---
    print("\n[10] Security Tests")

    # No auth
    r = requests.get(f"{INV_URL}/api/v1/equipment")
    test("No auth -> 401", r.status_code == 401, f"status={r.status_code}")

    r = requests.post(f"{INV_URL}/api/v1/equipment", json={"name": "hack"})
    test("No auth POST -> 401", r.status_code == 401, f"status={r.status_code}")

    # Invalid token
    bad_h = {"Authorization": "Bearer invalid.token.here", "Content-Type": "application/json"}
    r = requests.get(f"{INV_URL}/api/v1/equipment", headers=bad_h)
    test("Invalid token -> 401", r.status_code == 401, f"status={r.status_code}")

    # Invalid UUID in path
    r = requests.get(f"{INV_URL}/api/v1/equipment/not-a-uuid", headers=h)
    test("Invalid UUID -> 400", r.status_code == 400, f"status={r.status_code}")

    # Non-existent resource
    r = requests.get(f"{INV_URL}/api/v1/equipment/00000000-0000-0000-0000-000000000001", headers=h)
    test("Non-existent ID -> 404", r.status_code == 404, f"status={r.status_code}")

    # --- Cleanup ---
    print("\n[11] Cleanup (Delete)")
    if eq_id:
        r = requests.delete(f"{INV_URL}/api/v1/equipment/{eq_id}", headers=h)
        test("Delete equipment", r.status_code == 204, f"status={r.status_code}")

    if fc_id:
        r = requests.delete(f"{INV_URL}/api/v1/flightcases/{fc_id}", headers=h)
        test("Delete flightcase", r.status_code == 204, f"status={r.status_code}")

    if type_id:
        r = requests.delete(f"{INV_URL}/api/v1/equipment-types/{type_id}", headers=h)
        test("Delete equipment type", r.status_code == 204, f"status={r.status_code}")

    if cat2_id:
        r = requests.delete(f"{INV_URL}/api/v1/categories/{cat2_id}", headers=h)
        test("Delete category (Licht)", r.status_code == 204, f"status={r.status_code}")

    # Can't delete Audio category if it still has equipment from bulk create
    # So we skip that or accept failure
    if cat_id:
        r = requests.delete(f"{INV_URL}/api/v1/categories/{cat_id}", headers=h)
        if r.status_code == 204:
            test("Delete category (Audio)", True)
        else:
            test("Delete category blocked (has equipment)", r.status_code in (400, 409, 422),
                 f"status={r.status_code}")

    # --- Summary ---
    print("\n" + "=" * 60)
    print(f"Results: {passed} passed, {failed} failed, {passed + failed} total")
    if errors:
        print(f"\nFailed tests:")
        for e in errors:
            print(f"  - {e}")
    print("=" * 60)

    sys.exit(0 if failed == 0 else 1)


if __name__ == "__main__":
    main()
