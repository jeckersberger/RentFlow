#!/usr/bin/env python3
"""End-to-end integration test: cross-service flows for CrateDesk."""

import urllib.request
import json
import sys
import time

SERVICES = {
    "auth": "http://localhost:8001",
    "inventory": "http://localhost:8002",
    "project": "http://localhost:8003",
    "scanner": "http://localhost:8004",
    "warehouse": "http://localhost:8005",
    "invoice": "http://localhost:8006",
    "document": "http://localhost:8007",
    "crew": "http://localhost:8008",
    "federation": "http://localhost:8009",
    "maintenance": "http://localhost:8010",
    "transport": "http://localhost:8011",
    "insurance": "http://localhost:8012",
    "workflow": "http://localhost:8013",
    "ai": "http://localhost:8014",
    "notification": "http://localhost:8015",
    "reporting": "http://localhost:8016",
    "audit": "http://localhost:8017",
    "expense": "http://localhost:8018",
    "customer": "http://localhost:8019",
}

TOKEN = None
results = []


def req(base, method, path, body=None, auth=True):
    hdrs = {"Content-Type": "application/json"}
    if auth and TOKEN:
        hdrs["Authorization"] = f"Bearer {TOKEN}"
    data = json.dumps(body).encode() if body else None
    r = urllib.request.Request(base + path, data=data, headers=hdrs, method=method)
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
    except urllib.error.URLError:
        return 0, {"error": "connection refused"}


def test(name, ok, detail=""):
    status = "PASS" if ok else "FAIL"
    results.append((name, ok))
    msg = f"  [{status}] {name}"
    if detail and not ok:
        msg += f" -- {detail}"
    print(msg)


def get_data(resp):
    """Extract data from API envelope."""
    if isinstance(resp, dict):
        return resp.get("data", resp)
    return resp


# ── Phase 1: Health checks for all 19 services ──
print("\n=== Phase 1: Health Checks (all 19 services) ===")
healthy = 0
for name, base in SERVICES.items():
    code, _ = req(base, "GET", "/health", auth=False)
    ok = code == 200
    test(f"health:{name}", ok, f"status={code}")
    if ok:
        healthy += 1

if healthy == 0:
    print("\nNo services reachable. Start with: docker compose up -d")
    sys.exit(1)

print(f"\n  {healthy}/19 services healthy")

# ── Phase 2: Auth flow ──
print("\n=== Phase 2: Authentication ===")
code, body = req(SERVICES["auth"], "POST", "/api/v1/auth/login", {
    "email": "j.eckersberger@je-soundulight.de",
    "password": "RentFlow2026!",
    "tenant_slug": "je-soundulight",
}, auth=False)
login_ok = code == 200
data = get_data(body)
if login_ok and isinstance(data, dict):
    TOKEN = data.get("access_token") or data.get("token", "")
test("auth:login", login_ok and bool(TOKEN), f"code={code}")

# ── Phase 3: CRUD flow — Equipment ──
print("\n=== Phase 3: Equipment CRUD ===")
equip_id = None
code, body = req(SERVICES["inventory"], "POST", "/api/v1/equipment", {
    "name": "E2E Test Speaker",
    "sku": f"E2E-SPK-{int(time.time())}",
    "category_name": "Lautsprecher",
    "status": "available",
    "quantity": 4,
    "daily_rate": 5000,
    "replacement_value": 120000,
})
data = get_data(body)
if code == 201 and isinstance(data, dict):
    equip_id = data.get("id")
test("inventory:create_equipment", code == 201 and equip_id is not None, f"code={code}")

if equip_id:
    code, body = req(SERVICES["inventory"], "GET", f"/api/v1/equipment/{equip_id}")
    test("inventory:get_equipment", code == 200, f"code={code}")

    code, body = req(SERVICES["inventory"], "GET", "/api/v1/equipment?page=1&per_page=10")
    test("inventory:list_equipment", code == 200, f"code={code}")

# ── Phase 4: Customer + Project + Invoice flow ──
print("\n=== Phase 4: Customer -> Project -> Invoice ===")
cust_id = None
code, body = req(SERVICES["customer"], "POST", "/api/v1/customers", {
    "name": "E2E Testfirma GmbH",
    "customer_number": f"E2E-{int(time.time())}",
    "email": "test@e2e.de",
    "phone": "+49 123 456789",
    "address_street": "Teststrasse 1",
    "address_zip": "12345",
    "address_city": "Teststadt",
})
data = get_data(body)
if code == 201 and isinstance(data, dict):
    cust_id = data.get("id")
test("customer:create", code == 201 and cust_id is not None, f"code={code}")

proj_id = None
code, body = req(SERVICES["project"], "POST", "/api/v1/projects", {
    "name": "E2E Testprojekt",
    "status": "draft",
    "start_date": "2026-04-01",
    "end_date": "2026-04-03",
    "venue": "Stadthalle Teststadt",
    "customer_id": cust_id or "00000000-0000-0000-0000-000000000000",
})
data = get_data(body)
if code == 201 and isinstance(data, dict):
    proj_id = data.get("id")
test("project:create", code == 201 and proj_id is not None, f"code={code}")

inv_id = None
code, body = req(SERVICES["invoice"], "POST", "/api/v1/invoices", {
    "type": "invoice",
    "customer_name": "E2E Testfirma GmbH",
    "customer_email": "test@e2e.de",
    "vat_rate": 1900,
})
data = get_data(body)
if code == 201 and isinstance(data, dict):
    inv_id = data.get("id")
test("invoice:create", code == 201 and inv_id is not None, f"code={code}")

if inv_id:
    code, body = req(SERVICES["invoice"], "POST", f"/api/v1/invoices/{inv_id}/items", {
        "description": "E2E Test Speaker (4 Tage)",
        "quantity": 4,
        "unit_price": 5000,
        "position": 1,
    })
    test("invoice:add_item", code == 201, f"code={code}")

    code, body = req(SERVICES["invoice"], "GET", f"/api/v1/invoices/{inv_id}/items")
    test("invoice:list_items", code == 200, f"code={code}")

# ── Phase 5: Scanner flow ──
print("\n=== Phase 5: Scanner ===")
code, body = req(SERVICES["scanner"], "POST", "/api/v1/scanner/devices/register", {
    "device_id": f"e2e-device-{int(time.time())}",
    "model": "CF-H906",
    "firmware_version": "1.0.0",
})
test("scanner:register_device", code in (200, 201), f"code={code}")

code, body = req(SERVICES["scanner"], "POST", "/api/v1/scanner/scan", {
    "barcode": f"E2E-SPK-{int(time.time())}",
    "action": "scan",
    "device_id": f"e2e-device-{int(time.time())}",
})
test("scanner:scan_barcode", code in (200, 201), f"code={code}")

# ── Phase 6: Warehouse flow ──
print("\n=== Phase 6: Warehouse ===")
wh_id = None
code, body = req(SERVICES["warehouse"], "POST", "/api/v1/warehouses", {
    "name": "E2E Hauptlager",
    "code": f"E2E-WH-{int(time.time())}",
    "address": "Lagerstrasse 1, 12345 Teststadt",
})
data = get_data(body)
if code == 201 and isinstance(data, dict):
    wh_id = data.get("id")
test("warehouse:create", code == 201, f"code={code}")

# ── Phase 7: Crew flow ──
print("\n=== Phase 7: Crew ===")
crew_id = None
code, body = req(SERVICES["crew"], "POST", "/api/v1/crew", {
    "first_name": "Max",
    "last_name": "E2E-Tester",
    "email": "max@e2e.de",
    "phone": "+49 111 222333",
    "role": "technician",
})
data = get_data(body)
if code == 201 and isinstance(data, dict):
    crew_id = data.get("id")
test("crew:create_member", code == 201, f"code={code}")

if crew_id:
    code, body = req(SERVICES["crew"], "GET", f"/api/v1/crew/{crew_id}")
    test("crew:get_member", code == 200, f"code={code}")

# ── Phase 8: Maintenance flow ──
print("\n=== Phase 8: Maintenance ===")
if equip_id:
    code, body = req(SERVICES["maintenance"], "POST", "/api/v1/maintenance-tasks", {
        "equipment_id": equip_id,
        "title": "E2E Routine-Check",
        "description": "Automatischer E2E Test",
        "priority": "normal",
    })
    test("maintenance:create_task", code == 201, f"code={code}")

# ── Phase 9: Transport flow ──
print("\n=== Phase 9: Transport ===")
code, body = req(SERVICES["transport"], "POST", "/api/v1/vehicles", {
    "name": "E2E Transporter",
    "license_plate": f"E2E-{int(time.time()) % 10000}",
    "type": "van",
})
test("transport:create_vehicle", code == 201, f"code={code}")

# ── Phase 10: Expense flow ──
print("\n=== Phase 10: Expenses ===")
cat_id = None
code, body = req(SERVICES["expense"], "POST", "/api/v1/expense-categories", {
    "name": "E2E Reisekosten",
    "code": f"E2E-RK-{int(time.time())}",
})
data = get_data(body)
if code == 201 and isinstance(data, dict):
    cat_id = data.get("id")
test("expense:create_category", code == 201, f"code={code}")

if cat_id:
    code, body = req(SERVICES["expense"], "POST", "/api/v1/expenses", {
        "category_id": cat_id,
        "description": "E2E Testausgabe",
        "amount": 4999,
        "date": "2026-03-27",
    })
    test("expense:create_expense", code == 201, f"code={code}")

# ── Phase 11: Notification flow ──
print("\n=== Phase 11: Notifications ===")
code, body = req(SERVICES["notification"], "GET", "/api/v1/notifications/unread-count")
test("notification:unread_count", code == 200, f"code={code}")

# ── Phase 12: Audit flow ──
print("\n=== Phase 12: Audit ===")
code, body = req(SERVICES["audit"], "POST", "/api/v1/audit-logs", {
    "entity_type": "equipment",
    "entity_id": equip_id or "00000000-0000-0000-0000-000000000000",
    "action": "e2e_test",
    "description": "E2E integration test audit entry",
})
test("audit:create_log", code == 201, f"code={code}")

code, body = req(SERVICES["audit"], "GET", "/api/v1/audit-logs?page=1&per_page=5")
test("audit:list_logs", code == 200, f"code={code}")

# ── Phase 13: Workflow flow ──
print("\n=== Phase 13: Workflow ===")
wf_def_id = None
code, body = req(SERVICES["workflow"], "POST", "/api/v1/workflow-definitions", {
    "name": "E2E Genehmigung",
    "type": "approval",
    "steps": [{"name": "Pruefer", "role": "admin"}],
})
data = get_data(body)
if code == 201 and isinstance(data, dict):
    wf_def_id = data.get("id")
test("workflow:create_definition", code == 201, f"code={code}")

# ── Phase 14: Document flow ──
print("\n=== Phase 14: Documents ===")
code, body = req(SERVICES["document"], "POST", "/api/v1/document-templates", {
    "name": "E2E Lieferschein",
    "type": "delivery_note",
    "content": "<h1>Lieferschein</h1><p>{{customer_name}}</p>",
})
test("document:create_template", code == 201, f"code={code}")

# ── Phase 15: Reporting flow ──
print("\n=== Phase 15: Reporting ===")
code, body = req(SERVICES["reporting"], "POST", "/api/v1/report-definitions", {
    "name": "E2E Umsatzbericht",
    "type": "revenue",
    "query": "SELECT 1",
})
test("reporting:create_definition", code == 201, f"code={code}")

# ── Phase 16: Insurance flow ──
print("\n=== Phase 16: Insurance ===")
code, body = req(SERVICES["insurance"], "POST", "/api/v1/insurance-policies", {
    "name": "E2E Geraeteversicherung",
    "provider": "E2E Insurance GmbH",
    "policy_number": f"E2E-POL-{int(time.time())}",
    "premium": 50000,
    "coverage_amount": 5000000,
    "start_date": "2026-01-01",
    "end_date": "2026-12-31",
})
test("insurance:create_policy", code == 201, f"code={code}")

# ── Phase 17: Federation flow ──
print("\n=== Phase 17: Federation ===")
code, body = req(SERVICES["federation"], "POST", "/api/v1/federation-partners", {
    "name": "E2E Partner-Verleiher",
    "api_url": "https://partner.e2e.test/api",
    "status": "pending",
})
test("federation:create_partner", code == 201, f"code={code}")

# ── Phase 18: AI flow ──
print("\n=== Phase 18: AI ===")
code, body = req(SERVICES["ai"], "POST", "/api/v1/ai-predictions", {
    "type": "demand_forecast",
    "input_data": {"equipment_id": equip_id or "test", "period": "2026-Q2"},
    "prediction": {"expected_demand": 15},
    "confidence": 0.85,
})
test("ai:create_prediction", code == 201, f"code={code}")


# ── Summary ──
print("\n" + "=" * 60)
passed = sum(1 for _, ok in results if ok)
total = len(results)
failed = total - passed
print(f"  E2E Results: {passed}/{total} passed, {failed} failed")
if failed:
    print("\n  Failed tests:")
    for name, ok in results:
        if not ok:
            print(f"    - {name}")
print("=" * 60)
sys.exit(0 if failed == 0 else 1)
