#!/usr/bin/env python3
"""
CrateDesk Customer Service — Functional & Security Tests
Tests all CRUD operations for Customers, Contacts, and Contact Notes.
"""

import json
import sys
import uuid
import requests

AUTH_URL = "http://localhost:8001"
CUST_URL = "http://localhost:8019"
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
    print("CrateDesk Customer Service — Test Suite")
    print("=" * 60)

    uid = uuid.uuid4().hex[:6]

    # --- Auth ---
    print("\n[1] Authentication")
    token = login()
    test("Login successful", token != "", "No token returned")
    h = headers(token)

    # --- Health ---
    print("\n[2] Health Checks")
    r = requests.get(f"{CUST_URL}/health")
    test("Health endpoint", r.status_code == 200)
    r = requests.get(f"{CUST_URL}/livez")
    test("Liveness endpoint", r.status_code == 200)

    # --- Customers CRUD ---
    print("\n[3] Customers CRUD")
    r = requests.post(f"{CUST_URL}/api/v1/customers", headers=h, json={
        "company_name": f"Stadtwerke Muenchen-{uid}",
        "customer_number": f"KD-{uid}",
        "email": "info@stadtwerke-muc.de",
        "phone": "+49 89 12345678",
        "website": "https://www.stadtwerke-muc.de",
        "billing_address_street": "Marienplatz 1",
        "billing_address_city": "Muenchen",
        "billing_address_zip": "80331",
        "billing_address_country": "Deutschland",
        "tax_id": "DE123456789",
        "notes": "Grossveranstaltungen"
    })
    test("Create customer", r.status_code == 201, f"status={r.status_code} body={r.text[:300]}")
    cust_id = r.json().get("data", {}).get("id") if r.status_code == 201 else None

    r = requests.post(f"{CUST_URL}/api/v1/customers", headers=h, json={
        "company_name": f"Eventagentur Meier-{uid}",
        "email": "info@meier-events.de"
    })
    test("Create second customer", r.status_code == 201, f"status={r.status_code}")
    cust2_id = r.json().get("data", {}).get("id") if r.status_code == 201 else None

    r = requests.get(f"{CUST_URL}/api/v1/customers?page=1&per_page=50", headers=h)
    test("List customers", r.status_code == 200 and len(r.json().get("data", [])) >= 2,
         f"status={r.status_code} count={len(r.json().get('data', []))}")

    if cust_id:
        r = requests.get(f"{CUST_URL}/api/v1/customers/{cust_id}", headers=h)
        test("Get customer by ID", r.status_code == 200 and f"Stadtwerke" in r.json()["data"]["company_name"])

        r = requests.put(f"{CUST_URL}/api/v1/customers/{cust_id}", headers=h, json={
            "company_name": f"Stadtwerke Muenchen AG-{uid}",
            "notes": "VIP-Kunde"
        })
        test("Update customer", r.status_code == 200, f"status={r.status_code} body={r.text[:200]}")

    # --- Search ---
    print("\n[4] Search")
    r = requests.get(f"{CUST_URL}/api/v1/customers/search?q=Stadtwerke", headers=h)
    test("Search for 'Stadtwerke'", r.status_code == 200, f"status={r.status_code}")

    # --- Contacts CRUD ---
    print("\n[5] Contacts CRUD")
    if cust_id:
        r = requests.post(f"{CUST_URL}/api/v1/contacts", headers=h, json={
            "customer_id": cust_id,
            "first_name": "Thomas",
            "last_name": f"Mueller-{uid}",
            "email": "t.mueller@stadtwerke-muc.de",
            "phone": "+49 89 11111111",
            "mobile": "+49 171 1234567",
            "position": "Eventmanager",
            "is_primary": True
        })
        test("Create contact", r.status_code == 201, f"status={r.status_code} body={r.text[:200]}")
        contact_id = r.json().get("data", {}).get("id") if r.status_code == 201 else None

        r = requests.post(f"{CUST_URL}/api/v1/contacts", headers=h, json={
            "customer_id": cust_id,
            "first_name": "Anna",
            "last_name": f"Schmidt-{uid}",
            "email": "a.schmidt@stadtwerke-muc.de",
            "position": "Technik"
        })
        test("Create second contact", r.status_code == 201)

        r = requests.get(f"{CUST_URL}/api/v1/customers/{cust_id}/contacts", headers=h)
        test("List contacts for customer", r.status_code == 200 and len(r.json().get("data", [])) >= 2,
             f"status={r.status_code}")

        if contact_id:
            r = requests.get(f"{CUST_URL}/api/v1/contacts/{contact_id}", headers=h)
            test("Get contact by ID", r.status_code == 200 and "Thomas" in r.json()["data"]["first_name"])

            r = requests.put(f"{CUST_URL}/api/v1/contacts/{contact_id}", headers=h, json={
                "position": "Senior Eventmanager"
            })
            test("Update contact", r.status_code == 200, f"status={r.status_code} body={r.text[:200]}")

            # --- Contact Notes ---
            print("\n[6] Contact Notes")
            r = requests.post(f"{CUST_URL}/api/v1/contacts/{contact_id}/notes", headers=h, json={
                "content": "Telefonat am 15.03 — Budget fuer Sommerfest besprochen"
            })
            test("Add contact note", r.status_code == 201, f"status={r.status_code} body={r.text[:200]}")

            r = requests.post(f"{CUST_URL}/api/v1/contacts/{contact_id}/notes", headers=h, json={
                "content": "Angebot gesendet per Email"
            })
            test("Add second note", r.status_code == 201)

            r = requests.get(f"{CUST_URL}/api/v1/contacts/{contact_id}/notes", headers=h)
            test("Get contact notes", r.status_code == 200 and len(r.json().get("data", [])) >= 2,
                 f"status={r.status_code}")

            # Delete contact
            r = requests.delete(f"{CUST_URL}/api/v1/contacts/{contact_id}", headers=h)
            test("Delete contact (cascade notes)", r.status_code == 204, f"status={r.status_code}")

    # --- Security Tests ---
    print("\n[7] Security Tests")

    r = requests.get(f"{CUST_URL}/api/v1/customers")
    test("No auth -> 401", r.status_code == 401, f"status={r.status_code}")

    r = requests.post(f"{CUST_URL}/api/v1/customers", json={"company_name": "hack"})
    test("No auth POST -> 401", r.status_code == 401, f"status={r.status_code}")

    bad_h = {"Authorization": "Bearer invalid.token.here", "Content-Type": "application/json"}
    r = requests.get(f"{CUST_URL}/api/v1/customers", headers=bad_h)
    test("Invalid token -> 401", r.status_code == 401, f"status={r.status_code}")

    r = requests.get(f"{CUST_URL}/api/v1/customers/not-a-uuid", headers=h)
    test("Invalid UUID -> 400", r.status_code == 400, f"status={r.status_code}")

    r = requests.get(f"{CUST_URL}/api/v1/customers/00000000-0000-0000-0000-000000000001", headers=h)
    test("Non-existent ID -> 404", r.status_code == 404, f"status={r.status_code}")

    # --- Cleanup ---
    print("\n[8] Cleanup (Delete/Deactivate)")

    if cust2_id:
        r = requests.delete(f"{CUST_URL}/api/v1/customers/{cust2_id}", headers=h)
        test("Soft-delete customer 2", r.status_code == 204, f"status={r.status_code}")

    if cust_id:
        r = requests.delete(f"{CUST_URL}/api/v1/customers/{cust_id}", headers=h)
        test("Soft-delete customer 1", r.status_code == 204, f"status={r.status_code}")

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
