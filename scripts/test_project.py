#!/usr/bin/env python3
"""
CrateDesk Project Service — Functional & Security Tests
Tests all CRUD operations for Projects, Project Equipment, Packlists, and Reservations.
"""

import json
import sys
import uuid
import requests

AUTH_URL = "http://localhost:8001"
PROJ_URL = "http://localhost:8003"
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
    print("CrateDesk Project Service — Test Suite")
    print("=" * 60)

    uid = uuid.uuid4().hex[:6]

    # --- Auth ---
    print("\n[1] Authentication")
    token = login()
    test("Login successful", token != "", "No token returned")
    h = headers(token)

    # --- Health ---
    print("\n[2] Health Checks")
    r = requests.get(f"{PROJ_URL}/health")
    test("Health endpoint", r.status_code == 200)
    r = requests.get(f"{PROJ_URL}/livez")
    test("Liveness endpoint", r.status_code == 200)

    # --- Projects CRUD ---
    print("\n[3] Projects CRUD")
    r = requests.post(f"{PROJ_URL}/api/v1/projects", headers=h, json={
        "name": f"Sommerfest-{uid}",
        "project_number": f"PRJ-{uid}",
        "description": "Open-Air Festival",
        "contact_name": "Max Mustermann",
        "contact_email": "max@example.com",
        "venue_name": "Stadtpark",
        "venue_address": "Hauptstr. 1, 80331 Muenchen",
        "start_date": "2026-07-15",
        "end_date": "2026-07-17",
        "setup_date": "2026-07-14",
        "teardown_date": "2026-07-18",
        "budget": 500000,
        "notes": "VIP-Bereich beachten"
    })
    test("Create project", r.status_code == 201, f"status={r.status_code} body={r.text[:300]}")
    proj_id = r.json().get("data", {}).get("id") if r.status_code == 201 else None

    r = requests.post(f"{PROJ_URL}/api/v1/projects", headers=h, json={
        "name": f"Firmenfeier-{uid}",
        "start_date": "2026-08-01",
        "end_date": "2026-08-02"
    })
    test("Create second project", r.status_code == 201, f"status={r.status_code}")
    proj2_id = r.json().get("data", {}).get("id") if r.status_code == 201 else None

    r = requests.get(f"{PROJ_URL}/api/v1/projects?page=1&per_page=50", headers=h)
    test("List projects", r.status_code == 200 and len(r.json().get("data", [])) >= 2,
         f"status={r.status_code} count={len(r.json().get('data', []))}")

    if proj_id:
        r = requests.get(f"{PROJ_URL}/api/v1/projects/{proj_id}", headers=h)
        test("Get project by ID", r.status_code == 200 and f"Sommerfest-{uid}" in r.json()["data"]["name"])

        r = requests.put(f"{PROJ_URL}/api/v1/projects/{proj_id}", headers=h, json={
            "name": f"Sommerfest Updated-{uid}",
            "budget": 600000
        })
        test("Update project", r.status_code == 200, f"status={r.status_code} body={r.text[:200]}")

        # Status update
        r = requests.patch(f"{PROJ_URL}/api/v1/projects/{proj_id}/status", headers=h, json={
            "status": "confirmed"
        })
        test("Update status to confirmed", r.status_code == 204, f"status={r.status_code}")

        # Invalid status
        r = requests.patch(f"{PROJ_URL}/api/v1/projects/{proj_id}/status", headers=h, json={
            "status": "bogus"
        })
        test("Invalid status rejected", r.status_code in (400, 422), f"status={r.status_code}")

    # --- Search ---
    print("\n[4] Search")
    r = requests.get(f"{PROJ_URL}/api/v1/projects/search?q=Sommerfest", headers=h)
    test("Search for 'Sommerfest'", r.status_code == 200, f"status={r.status_code}")

    # --- Project Equipment ---
    print("\n[5] Project Equipment")
    fake_equipment_id = str(uuid.uuid4())

    if proj_id:
        r = requests.post(f"{PROJ_URL}/api/v1/projects/{proj_id}/equipment", headers=h, json={
            "equipment_id": fake_equipment_id,
            "quantity": 4,
            "allocated_from": "2026-07-14",
            "allocated_until": "2026-07-18",
            "notes": "Hauptbuehne links"
        })
        test("Add equipment to project", r.status_code == 201, f"status={r.status_code} body={r.text[:200]}")
        pe_id = r.json().get("data", {}).get("id") if r.status_code == 201 else None

        r = requests.get(f"{PROJ_URL}/api/v1/projects/{proj_id}/equipment", headers=h)
        test("List project equipment", r.status_code == 200 and len(r.json().get("data", [])) >= 1,
             f"status={r.status_code}")

        if pe_id:
            r = requests.delete(f"{PROJ_URL}/api/v1/projects/{proj_id}/equipment/{pe_id}", headers=h)
            test("Remove equipment from project", r.status_code == 204, f"status={r.status_code}")

    # --- Packlists ---
    print("\n[6] Packlists")

    if proj_id:
        r = requests.post(f"{PROJ_URL}/api/v1/packlists", headers=h, json={
            "project_id": proj_id,
            "name": f"Packliste Audio-{uid}"
        })
        test("Create packlist", r.status_code == 201, f"status={r.status_code} body={r.text[:200]}")
        pl_id = r.json().get("data", {}).get("id") if r.status_code == 201 else None

        if pl_id:
            r = requests.get(f"{PROJ_URL}/api/v1/packlists/{pl_id}", headers=h)
            test("Get packlist by ID", r.status_code == 200)

            # Add item
            r = requests.post(f"{PROJ_URL}/api/v1/packlists/{pl_id}/items", headers=h, json={
                "equipment_id": fake_equipment_id,
                "quantity_planned": 2,
                "notes": "Mikrofone"
            })
            test("Add packlist item", r.status_code == 201, f"status={r.status_code} body={r.text[:200]}")
            item_id = r.json().get("data", {}).get("id") if r.status_code == 201 else None

            # Get items
            r = requests.get(f"{PROJ_URL}/api/v1/packlists/{pl_id}/items", headers=h)
            test("Get packlist items", r.status_code == 200 and len(r.json().get("data", [])) >= 1,
                 f"status={r.status_code}")

            # Update packed
            if item_id:
                r = requests.patch(f"{PROJ_URL}/api/v1/packlists/{pl_id}/items/{item_id}/packed", headers=h, json={
                    "quantity_packed": 2
                })
                test("Update item packed", r.status_code == 204, f"status={r.status_code}")

                r = requests.patch(f"{PROJ_URL}/api/v1/packlists/{pl_id}/items/{item_id}/returned", headers=h, json={
                    "quantity_returned": 2
                })
                test("Update item returned", r.status_code == 204, f"status={r.status_code}")

            # Update status
            r = requests.patch(f"{PROJ_URL}/api/v1/packlists/{pl_id}/status", headers=h, json={
                "status": "packing"
            })
            test("Update packlist status", r.status_code == 204, f"status={r.status_code}")

        # List by project
        r = requests.get(f"{PROJ_URL}/api/v1/projects/{proj_id}/packlists", headers=h)
        test("List packlists by project", r.status_code == 200, f"status={r.status_code}")

    # --- Reservations ---
    print("\n[7] Reservations")

    if proj_id:
        r = requests.post(f"{PROJ_URL}/api/v1/reservations", headers=h, json={
            "project_id": proj_id,
            "equipment_id": fake_equipment_id,
            "quantity": 2,
            "start_date": "2026-07-14",
            "end_date": "2026-07-18"
        })
        test("Create reservation", r.status_code == 201, f"status={r.status_code} body={r.text[:200]}")
        res_id = r.json().get("data", {}).get("id") if r.status_code == 201 else None

        r = requests.get(f"{PROJ_URL}/api/v1/reservations", headers=h)
        test("List reservations", r.status_code == 200, f"status={r.status_code}")

        if res_id:
            r = requests.get(f"{PROJ_URL}/api/v1/reservations/{res_id}", headers=h)
            test("Get reservation by ID", r.status_code == 200)

            r = requests.patch(f"{PROJ_URL}/api/v1/reservations/{res_id}/status", headers=h, json={
                "status": "confirmed"
            })
            test("Update reservation status", r.status_code == 204, f"status={r.status_code}")

    # --- Security Tests ---
    print("\n[8] Security Tests")

    r = requests.get(f"{PROJ_URL}/api/v1/projects")
    test("No auth -> 401", r.status_code == 401, f"status={r.status_code}")

    r = requests.post(f"{PROJ_URL}/api/v1/projects", json={"name": "hack"})
    test("No auth POST -> 401", r.status_code == 401, f"status={r.status_code}")

    bad_h = {"Authorization": "Bearer invalid.token.here", "Content-Type": "application/json"}
    r = requests.get(f"{PROJ_URL}/api/v1/projects", headers=bad_h)
    test("Invalid token -> 401", r.status_code == 401, f"status={r.status_code}")

    r = requests.get(f"{PROJ_URL}/api/v1/projects/not-a-uuid", headers=h)
    test("Invalid UUID -> 400", r.status_code == 400, f"status={r.status_code}")

    r = requests.get(f"{PROJ_URL}/api/v1/projects/00000000-0000-0000-0000-000000000001", headers=h)
    test("Non-existent ID -> 404", r.status_code == 404, f"status={r.status_code}")

    # --- Cleanup ---
    print("\n[9] Cleanup (Delete)")

    if proj2_id:
        r = requests.delete(f"{PROJ_URL}/api/v1/projects/{proj2_id}", headers=h)
        test("Delete second project", r.status_code == 204, f"status={r.status_code}")

    if proj_id:
        r = requests.delete(f"{PROJ_URL}/api/v1/projects/{proj_id}", headers=h)
        test("Delete project (cascade)", r.status_code == 204, f"status={r.status_code}")

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
