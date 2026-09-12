# ADR-001: Rental core data ownership

Status: Accepted for implementation

## Context

CrateDesk currently spreads project planning, inventory state, scanner actions and warehouse movements across multiple services. Several of those services can represent related facts, which makes it possible for two write paths to disagree about whether material is required, reserved, physically present or already fulfilled.

Before the availability and migration work continues, every operational fact needs one authoritative writer.

## Decision

The existing `inventory-service` will evolve into the authoritative **rental core** for operational material state. This is an incremental change; it is not a rewrite of all services and does not require moving unrelated project, invoice or document data into the inventory database.

### Authoritative ownership

| Fact | Authoritative writer | Notes |
| --- | --- | --- |
| project/customer master data | `project-service` | Names, dates, customer references, descriptive project data |
| material requirement | rental core (`inventory-service`) | Versioned requirement revisions per project |
| material allocation / capacity commitment | rental core | The only write path allowed to consume availability |
| equipment blocking state | rental core | Damage, inspection hold, overdue/other operational blocks |
| physical material movement | rental core | Check-out, return, relocation and inventory adjustments |
| fulfilment state | rental core | Packed / checked-out / returned quantities derived from accepted commands |
| warehouse/location master data | `warehouse-service` during transition | Rental core references stable location IDs; write ownership may be revisited later |
| invoice/accounting state | `invoice-service` | Must consume operational facts; it must not mutate material state |
| documents | `document-service` | Stores generated/original documents; documents do not become authoritative material state |
| audit trail | audit/outbox infrastructure | Records committed actions; it does not replace the authoritative domain rows |

## Domain terms

### Requirement

A requirement says **what a project needs** for an effective time range. A requirement does not by itself reserve inventory.

Required minimum fields:

- tenant ID
- project ID
- requirement revision ID
- line ID
- item/equipment type reference
- requested quantity
- effective start/end including operational buffers
- requirement status

Only one revision is active for a project at a time. Editing project material creates a new revision instead of destructively replacing the previous committed state.

### Allocation

An allocation says **which capacity is committed** to a requirement.

For serialized equipment, a committed allocation references a concrete equipment unit. For interchangeable quantity pools, it references the pool/type/location capacity and a quantity.

Creating, replacing, resizing or cancelling a committed allocation is a capacity-changing operation and must use the rental core transaction/concurrency rules.

### Movement

A movement says **what physically happened**. Examples:

- warehouse -> customer/project (check-out)
- customer/project -> warehouse (return)
- location A -> location B
- stock correction after approved inventory count

A movement is not a second reservation. Checking out already allocated equipment changes custody/fulfilment state without subtracting the same capacity a second time.

### Fulfilment state

Fulfilment state describes how much of a requirement has been operationally processed. At minimum it distinguishes:

- required
- allocated
- packed
- checked out
- returned and accepted
- returned but blocked/damaged
- missing/open

These quantities may differ legitimately. The system must not collapse them into one generic `status` field.

### Block

A block makes capacity unavailable independently of project demand. Examples include damage, inspection hold and an overdue unit that has not physically returned.

Blocks have an explicit reason, effective period/state and provenance. Removing a block is an audited domain action.

## Invariants

The implementation must preserve the following invariants:

1. A requirement can be approved while partially uncovered; uncovered quantity remains explicit.
2. A committed allocation can only be created by the rental core.
3. The sum of committed quantity allocations must never exceed available pool capacity at any instant in the effective interval.
4. One serialized unit cannot have overlapping committed allocations.
5. Check-out does not consume capacity a second time when an allocation already exists.
6. Return does not silently make damaged or inspection-held equipment available.
7. A missing/partial return remains open until explicitly resolved.
8. Repeating the same accepted operation ID must not create another effective movement/allocation.
9. Tenant identity comes from verified request context; clients cannot select another tenant by payload/header convention.
10. Downstream document, notification or reporting failure must not roll back an already committed physical movement.

## Command boundary

Web, Android scanner and future Federation integration must ultimately call the same rental-core commands for material mutations. They may have different transport endpoints, but must not each implement independent stock arithmetic.

Initial command families:

- `CreateRequirementRevision`
- `CommitAllocation`
- `ReleaseAllocation`
- `ReplaceSerializedAllocation`
- `CheckOutMaterial`
- `ReturnMaterial`
- `RelocateMaterial`
- `ApplyEquipmentBlock`
- `ReleaseEquipmentBlock`
- `ApplyApprovedInventoryAdjustment`

Every mutating command will carry or receive a stable operation/idempotency ID before offline scanner support is considered production-ready.

## Read models

Services and UIs may maintain read models for convenience. A read model is disposable and must not become an alternate source of truth for capacity or physical state.

Examples:

- project material summary in `project-service`
- pack-list presentation model
- reporting aggregates
- scanner cached job/session view

If a read model disagrees with the rental core, the rental core wins and the read model is rebuilt/reconciled.

## Cross-service side effects

A rental-core database transaction may persist domain state plus an outbox record atomically. Email, PDF generation, project summaries and other remote side effects happen after commit and must be retryable/idempotent.

No implementation should pretend that writes to separate service databases form one ACID transaction.

## Transition rules

Until a legacy write path has been replaced and verified:

1. Do not delete its data blindly.
2. Add adapters/read-throughs where required for migration.
3. Introduce the new authoritative write path first.
4. Migrate/reconcile existing records with explicit counts and exception reports.
5. Disable the old writer.
6. Remove old tables/code only after production-like verification shows no remaining readers/writers.

Stable existing IDs, barcodes and externally referenced identifiers must be retained wherever possible.

## Acceptance examples

### Sequential bookings

Stock: 3 units. Existing booking A uses 2 from 08:00-10:00; booking B uses 2 from 10:00-12:00. A new request for 1 unit from 08:00-12:00 remains feasible because no instant exceeds capacity.

### Check-out

Five units are required and allocated. Checking out those five records custody and fulfilment but does not turn the effective demand into ten units.

### Partial return

Twelve serialized units leave. Ten return serviceable, one returns damaged, one is missing. After acceptance, ten may become available, the damaged unit remains blocked, and the missing unit remains physically outstanding and unavailable.

### Concurrent final unit

Two requests race for the last available unit. At most one transaction may commit a covered allocation; the other receives a conflict/shortage result rather than a second successful commitment.

## Consequences

### Positive

- one place enforces availability and physical-state invariants
- scanner and browser behavior can converge on the same commands
- future Federation becomes an integration over a proven local workflow rather than a second booking engine
- invoice/reporting logic consumes stable operational facts

### Cost

- existing service boundaries need adapters during transition
- some current tables/status fields will become read models or legacy data
- migrations need explicit reconciliation rather than simple schema replacement
- availability/allocation writes require deliberate transaction design

## Non-goals

This ADR does not define the final SQL schema, invoice model, Federation protocol or offline conflict UX. Those are separate implementation tasks constrained by the ownership and invariants above.
