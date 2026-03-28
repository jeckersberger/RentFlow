#!/usr/bin/env python3
"""
EquipFlow Invoice Service — Functional & Security Tests
Tests CRUD operations, items, finalize, payments, and GoBD compliance.
"""

import json
import sys
import uuid
import requests

AUTH_URL = "http://localhost:8001"
INV_URL = "http://localhost:8006"
EMAIL = "j.eckersberger@je-soundulight.de"
PASSWORD = "RentFlow2026!"
TENANT_SLUG = "je-soundulight"

passed = 0
failed = 0
errors_list = []


def test(name, condition, detail=""):
    global passed, failed
    if condition:
        passed += 1
        print(f"  [PASS] {name}")
    else:
        failed += 1
        errors_list.append(name)
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
    print("EquipFlow Invoice Service — Test Suite")
    print("=" * 60)

    uid = uuid.uuid4().hex[:6]

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

    # --- Invoice CRUD ---
    print("\n[3] Invoice CRUD")
    r = requests.post(f"{INV_URL}/api/v1/invoices", headers=h, json={
        "customer_name": f"Stadtwerke Muenchen-{uid}",
        "customer_email": "buchhaltung@stadtwerke-muc.de",
        "customer_address": "Marienplatz 1, 80331 Muenchen",
        "invoice_date": "2026-03-15",
        "due_date": "2026-04-14",
        "vat_rate": 1900,
        "notes": "Sommerfest-Technik"
    })
    test("Create invoice", r.status_code == 201, f"status={r.status_code} body={r.text[:300]}")
    inv_id = r.json().get("data", {}).get("id") if r.status_code == 201 else None
    inv_number = r.json().get("data", {}).get("invoice_number", "") if r.status_code == 201 else ""
    test("Invoice number generated", inv_number.startswith("RE-"), f"number={inv_number}")

    r = requests.post(f"{INV_URL}/api/v1/invoices", headers=h, json={
        "customer_name": f"Eventagentur Meier-{uid}",
        "kleinunternehmer": True
    })
    test("Create Kleinunternehmer invoice", r.status_code == 201, f"status={r.status_code}")
    inv2_id = r.json().get("data", {}).get("id") if r.status_code == 201 else None
    inv2_data = r.json().get("data", {}) if r.status_code == 201 else {}
    test("Kleinunternehmer VAT=0", inv2_data.get("vat_rate") == 0, f"vat_rate={inv2_data.get('vat_rate')}")

    r = requests.get(f"{INV_URL}/api/v1/invoices?page=1&per_page=50", headers=h)
    test("List invoices", r.status_code == 200 and len(r.json().get("data", [])) >= 2,
         f"status={r.status_code} count={len(r.json().get('data', []))}")

    if inv_id:
        r = requests.get(f"{INV_URL}/api/v1/invoices/{inv_id}", headers=h)
        test("Get invoice by ID", r.status_code == 200 and "Stadtwerke" in r.json()["data"]["customer_name"])

        r = requests.put(f"{INV_URL}/api/v1/invoices/{inv_id}", headers=h, json={
            "notes": "VIP-Kunde, Sonderkondition"
        })
        test("Update invoice (draft)", r.status_code == 200, f"status={r.status_code} body={r.text[:200]}")

    # --- Items ---
    print("\n[4] Invoice Items")
    if inv_id:
        r = requests.post(f"{INV_URL}/api/v1/invoices/{inv_id}/items", headers=h, json={
            "description": "L-Acoustics K2, 4x, 3 Tage",
            "quantity": 4,
            "unit": "Stueck",
            "unit_price": 25000,
            "position": 1
        })
        test("Add item 1", r.status_code == 201, f"status={r.status_code} body={r.text[:200]}")
        item1_id = r.json().get("data", {}).get("id") if r.status_code == 201 else None

        r = requests.post(f"{INV_URL}/api/v1/invoices/{inv_id}/items", headers=h, json={
            "description": "Tontechniker, 2 Tage",
            "quantity": 2,
            "unit": "Tag",
            "unit_price": 60000,
            "position": 2
        })
        test("Add item 2", r.status_code == 201)

        r = requests.get(f"{INV_URL}/api/v1/invoices/{inv_id}/items", headers=h)
        test("List items", r.status_code == 200 and len(r.json().get("data", [])) >= 2,
             f"status={r.status_code}")

    # --- Finalize (GoBD) ---
    print("\n[5] Finalize (GoBD)")
    if inv_id:
        r = requests.patch(f"{INV_URL}/api/v1/invoices/{inv_id}/finalize", headers=h)
        test("Finalize invoice", r.status_code == 200, f"status={r.status_code} body={r.text[:300]}")
        if r.status_code == 200:
            data = r.json()["data"]
            test("Status is finalized", data["status"] == "finalized")
            test("Hash generated", len(data.get("hash", "")) == 64, f"hash={data.get('hash', '')[:20]}")
            test("Total net calculated", data["total_net"] == 220000, f"net={data['total_net']}")
            # 220000 * 19% = 41800
            test("VAT calculated (19%)", data["total_vat"] == 41800, f"vat={data['total_vat']}")
            test("Total gross", data["total_gross"] == 261800, f"gross={data['total_gross']}")

        # Cannot edit finalized invoice
        r = requests.put(f"{INV_URL}/api/v1/invoices/{inv_id}", headers=h, json={
            "notes": "Trying to change"
        })
        test("Cannot edit finalized", r.status_code != 200, f"status={r.status_code}")

    # --- Payments ---
    print("\n[6] Payments")
    if inv_id:
        r = requests.post(f"{INV_URL}/api/v1/invoices/{inv_id}/payments", headers=h, json={
            "amount": 100000,
            "payment_date": "2026-04-01",
            "payment_method": "bank_transfer",
            "reference": "SEPA-2026-0001"
        })
        test("Add partial payment", r.status_code == 201, f"status={r.status_code} body={r.text[:200]}")

        r = requests.get(f"{INV_URL}/api/v1/invoices/{inv_id}", headers=h)
        if r.status_code == 200:
            test("Status partial_paid", r.json()["data"]["status"] == "partial_paid",
                 f"status={r.json()['data']['status']}")

        r = requests.post(f"{INV_URL}/api/v1/invoices/{inv_id}/payments", headers=h, json={
            "amount": 161800,
            "payment_date": "2026-04-10",
            "payment_method": "bank_transfer"
        })
        test("Add final payment", r.status_code == 201)

        r = requests.get(f"{INV_URL}/api/v1/invoices/{inv_id}", headers=h)
        if r.status_code == 200:
            test("Status paid", r.json()["data"]["status"] == "paid",
                 f"status={r.json()['data']['status']}")

        r = requests.get(f"{INV_URL}/api/v1/invoices/{inv_id}/payments", headers=h)
        test("List payments", r.status_code == 200 and len(r.json().get("data", [])) >= 2)

    # --- Search ---
    print("\n[7] Search")
    r = requests.get(f"{INV_URL}/api/v1/invoices/search?q=Stadtwerke", headers=h)
    test("Search invoices", r.status_code == 200, f"status={r.status_code}")

    # --- Security Tests ---
    print("\n[8] Security Tests")
    r = requests.get(f"{INV_URL}/api/v1/invoices")
    test("No auth -> 401", r.status_code == 401, f"status={r.status_code}")

    r = requests.post(f"{INV_URL}/api/v1/invoices", json={"customer_name": "hack"})
    test("No auth POST -> 401", r.status_code == 401, f"status={r.status_code}")

    bad_h = {"Authorization": "Bearer invalid.token.here", "Content-Type": "application/json"}
    r = requests.get(f"{INV_URL}/api/v1/invoices", headers=bad_h)
    test("Invalid token -> 401", r.status_code == 401, f"status={r.status_code}")

    r = requests.get(f"{INV_URL}/api/v1/invoices/not-a-uuid", headers=h)
    test("Invalid UUID -> 400", r.status_code == 400, f"status={r.status_code}")

    r = requests.get(f"{INV_URL}/api/v1/invoices/00000000-0000-0000-0000-000000000001", headers=h)
    test("Non-existent ID -> 404", r.status_code == 404, f"status={r.status_code}")

    # --- Summary ---
    print("\n" + "=" * 60)
    print(f"Results: {passed} passed, {failed} failed, {passed + failed} total")
    if errors_list:
        print(f"\nFailed tests:")
        for e in errors_list:
            print(f"  - {e}")
    print("=" * 60)

    sys.exit(0 if failed == 0 else 1)


if __name__ == "__main__":
    main()
