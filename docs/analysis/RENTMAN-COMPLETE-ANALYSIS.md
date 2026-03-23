# Rentman Complete Application Analysis

**Analysis Date:** 2026-03-23
**Instance:** jhons.rentmanapp.com
**App Version:** 5.1189.0.5
**Language:** German (Deutsch)
**Account:** Jhons (Trial - 28 days remaining)
**Logged-in User:** John Dee

---

## 1. Navigation & Information Architecture

### 1.1 Global Layout

Rentman uses a classic **three-zone layout**:

1. **Left Sidebar Navigation** -- collapsible (chevron toggle), contains all modules organized as expandable groups
2. **Top Bar** -- contains: open tabs, trial status banner, global search (`/` shortcut), availability quick lookup (`alt+/`), help, notifications bell, user menu (avatar + name)
3. **Main Content Area** -- occupies rest of viewport, uses a tab bar system for multiple open entities

### 1.2 Tab System

The top bar includes a **multi-tab system** (like browser tabs). Each opened entity (project, config page, planner, etc.) gets its own tab with:
- Icon indicating type (table_chart for projects, category for materials, etc.)
- Lock icon (lock/lock_open)
- Entity name
- Close button (highlight_off)
- A "close all non-current tabs" button

### 1.3 Navigation Tree (Complete)

```
Implementierungsleitfaden (fact_check)
Dashboard (dashboard)
(Mein) Kalender (event)
  ├─ (Mein) Kalender
  └─ Job Board
Lager (warehouse icon)
  ├─ Lager
  ├─ Kombinationen
  ├─ Cross-Docking-Übersicht
  └─ Lager-Tracking-Log
Projekte (table_chart)
  ├─ Projekte
  └─ Vermietungsanfrage
Personalplaner (account_circle)
Mangel (swap_horizontal_circle)
  ├─ Mietengpässe
  ├─ Verkaufsengpässe
  └─ Zumietungsjobs
Finanzen (monetization_on)
  ├─ Rechnungen
  ├─ Zu fakturieren
  └─ Bestellungen
Materialien (category)
  ├─ Materialien
  ├─ Seriennummern
  ├─ Lagerstandorte
  ├─ Archivierte Materialien
  └─ Archivierte Standorte
Kontakte (contact_phone)
Mitarbeiter (people)
Fahrzeuge (local_shipping)
Aufgaben (assignment_turned_in)
Stundenerfassung (watch_later)
  ├─ Stundenerfassung
  ├─ Aktivitäten
  └─ Abwesenheitsanträge
Werkstatt (build)
  ├─ Reparaturen
  ├─ Prüfungen
  ├─ Zu prüfende Materialien
  ├─ Verlorene Materialien
  └─ Bestandszählungen
Statistik (poll)
Kommunikation (dvr)
  ├─ Kommunikations-Log
  ├─ Gesendete E-Mails
  └─ Erhaltene Notizen
Konfiguration (settings)
```

**Total: 16 top-level modules, ~35 sub-pages**

### 1.4 URL Structure

All routes use hash-based navigation (`#/`):
- `#/dashboard/20` -- Dashboard (numbered dashboards)
- `#/projects/single` -- Projects list
- `#/projects/80/details` -- Project detail (ID=80)
- `#/projects/80/materials` -- Project material tab
- `#/projects/80/financial` -- Project finance tab
- `#/equipment/349` -- Material detail (ID=349)
- `#/warehouse` -- Warehouse
- `#/planner` -- Crew planner
- `#/subrent/shortages` -- Rental shortages
- `#/invoices` -- Invoices
- `#/invoices/purchase-orders` -- Purchase orders
- `#/contacts/list` -- Contacts
- `#/employees` -- Employees
- `#/transport` -- Vehicles
- `#/tasks` -- Tasks
- `#/timeregistration/list` -- Time tracking
- `#/statistics` -- Statistics
- `#/logscommunication` -- Communication log
- `#/configpanel/{section}` -- Configuration

---

## 2. Module-by-Module Analysis

### 2.1 Dashboard (`#/dashboard`)

**Title:** Dashboard | Rentman Vermietungssoftware

**Features:**
- Multiple named dashboards (switchable via dropdown): "Finanzen", "My Personal Dashboard"
- "Neues Dashboard" button to create additional dashboards
- "Layout anpassen" (edit/save layout) toggle
- "Widget zufugen" (add widget) button
- Help link to support article

**Widgets observed:**
1. **Offenstehende Rechnung (Outstanding Invoice)** -- Shows outstanding invoice amount excl. VAT, broken down into:
   - Offen (Open): amount
   - Uberfällig (Overdue): amount
   - Zu fakturieren (To be invoiced): amount
   - Bar chart visualization
   - Launch icon to navigate to detail

2. **Umsatz (Tortendiagramm) / Revenue Pie Chart** -- Revenue visualization by period, with info tooltip

3. **Offenes Angebot (Open Quotes)** -- Date-range filtered (e.g., "Bis nachsten Monat"), shows pending quotes

4. **Einnahmen (Balkendiagramm) / Revenue Bar Chart** -- Revenue bar chart with date picker, launch link

**Widget actions:** Each widget has:
- Launch icon (opens detail view)
- More options (more_vert menu)
- Some have date range filters

**UI Pattern:** Card-based grid layout, drag-to-rearrange when in edit mode

---

### 2.2 Projects (`#/projects/single`)

**Title:** Projekte | Rentman Vermietungssoftware

#### 2.2.1 Project List View

**Header actions:**
- "Projekt hinzufugen" (Add project) button
- Help link

**Filter bar:**
- Date range filter: "Nachste Woche" dropdown with arrow_drop_down
- Location filter: "Lagerhaus Ost" dropdown
- Search toggle (search/close)
- View mode: "Ubersicht: Liste" dropdown (list icon)
- Status filter: "Projektstatus" dropdown
- Tags filter: "Tags" dropdown
- Custom filter: "Filter" dropdown

**Grid columns observed:**
- Checkbox (select)
- Color indicator (Farbe)
- Number (Nummer)
- Name
- Project Status (Projektstatus)
- Sortable columns with "keine Sortierung" / "aufsteigend sortieren" indicators

**Status values observed:** Am Veranstaltungsort, Bestätigt, Gepackt, Option, Annulliert

**Bulk actions toolbar (appears on selection):**
- Close selection
- "X Positionen ausgewahlt" count
- Edit (Bearbeiten)
- Print/Create document (Projektdokument erstellen)
- Open timeline (Zeitleiste offnen)
- More options (more_vert)

**View customization:**
- "Ansichten" settings dropdown

**Row interaction:**
- Single click: shows inline "Details" button + more_vert menu
- Double click: opens full project detail page

**Quick Details Panel (right sidebar):**
When clicking a project row, a detail sidebar appears with:
- "Details Projekt" header with lock_open/lock toggle and close
- Project number + name (e.g., "80 - Liveauftritt Band Soultrain")
- Project type (Projekttyp: Band)
- Created date
- Account manager with avatar
- **Auftraggeber (Client)** section: Contact name, address, Google Maps link
- **Standort (Location)** section: Venue, address, Google Maps link
- **Projektfortschritt (Project Progress)** icons: Material, Personnel, Quotes, Finance, Tasks
- **Planung (Planning)**: Calculated planning period and usage period
- **Zumietung (Sub-rental)**: Sub-rental status
- **Gesamt-Projekt (Overall)**: Weight (kg), Volume (m3), Power (A), Total power (W)
- **Tags** section with add button
- **Aufgaben (Tasks)** with count, filter, add -- shows task items with status badges (Abgelaufen = Expired)
- **Notizen (Notes)** with count and add
- **Dateien (Files)** with count, shows table of project documents (.pdf files with names and dates)

#### 2.2.2 Project Detail View (`#/projects/{id}/details`)

**Header:**
- Project number + name as heading
- Subproject dropdown ("Keine Subprojekte")
- Close / Save buttons
- "Nicht gespeicherte Anderungen" warning when dirty

**Tab bar (11 tabs):**
1. **Allgemein** (General)
2. **Zeitplan** (Schedule)
3. **Material** (Equipment)
4. **Personal und Transport** (Crew & Transport)
5. **Zusatzkosten** (Additional costs)
6. **Finanzen** (Finances)
7. **Zumietung** (Sub-rental)
8. **Bestellungen** (Purchase orders)
9. **Personalplanung** (Crew planning)
10. **Transportplanung** (Transport planning)
11. **History Log**

#### Allgemein (General) Tab Fields:

**Projekt section:**
- Projektvorlage anwenden (Apply project template)
- Projektdokument erstellen (Create project document) with dropdown arrow
- Name (textbox)
- Projekttyp (Project type): e.g., "Band" with edit button
- Status: dropdown (Bestätigt, Option, Am Veranstaltungsort, Gepackt, Annulliert)
- Nummer (Number): optional textbox
- Accountmanager: dropdown (e.g., "John Dee")
- Farbe (Color): color picker with hex input (#14B4EC)
- Externe Referenz (External reference): textbox
- Lagerstandort (Warehouse location): dropdown (e.g., "Lagerhaus Ost")

**Auftraggeber (Client) section:**
- Contact selector with delete button
- Address display with Google Maps link
- Kontaktperson (Contact person) dropdown

**Standort (Location) section:**
- "Vom Kunden ubernehmen" (Copy from customer) button
- Contact/venue with address
- Kontaktperson dropdown
- Google Maps link with embedded map
- Strecke (Distance): e.g., 105.00 km with info tooltip
- Fahrtzeit (Travel time): e.g., 0 Minuten
- "Ubernehmen" (Apply) button with refresh icon

**Tasks/Notes/Files panel:**
- Tab selector: Aufgaben (1), Notizen, Dateien (2)
- Filter and Add buttons
- Task items with check/edit/delete actions
- Task status badges (Abgelaufen)
- Created by/date info

**Right sidebar:**
- **Projektzeitraum (Project Period)**: Calendar widget with month/year dropdowns, week navigation, highlighted date range
- **Berechneter Planungszeitraum (Calculated Planning Period)**: Date range display
- **Projektfortschritt (Project Progress)**: Checklist of progress items:
  - Planungsmaterial: "Alle Mietmaterialien sind reserviert"
  - Personal & Transport: Warning about unplanned functions
  - Angebote: "Projekt wurde bestätigt"
  - Rechnungen: "Keine Rechnungen"
  - Aufgaben: "Fristen abgelaufen"
- **Tags** section with edit button

#### Material Tab (`#/projects/{id}/materials`)

- Fullscreen toggle (crop_free)
- Search bar with tag filter
- Add items panel with search, timeline, and "Hinzufugen" button with chevron
- Material grid with hierarchical folder structure
- Barcode/QR scanning support implied

#### Finanzen (Finance) Tab (`#/projects/{id}/financial`)

**Summary Cards:**
- **In Rechnung gestellt (Invoiced)**: Progress ring (0%), amount / total (e.g., 0.00 EUR / 3,245.40 EUR)
- **Cost comparison cards:**
  - Geschätzte Kosten (Estimated costs) with profit %
  - Geplante Kosten (Planned costs) with profit %
  - Tatsächliche Kosten (Actual costs) with profit %
  - Rabatt (Discount): amount + percentage

**Angebote/Verträge (Quotes/Contracts):**
- Tab toggle: Angebote (1) / Verträge (0)
- "Angebot hinzufugen" button
- Table columns: Nummer/Version, Status, Aufrufe (Views), Fälligkeitsdatum (Due date), Preis exkl. MwSt., Actions
- Quote status values: Unveröffentlicht
- Actions: view count, more_vert menu

**Rechnungen (Invoices):**
- "Rechnung hinzufugen" button
- Empty state with illustration and helper text

**Finanzen Ubersicht (Financial Overview) Table:**
- Categories as rows: Vermietung (Rental), Verkauf (Sale), Personal (Crew), Transport, Zusatzkosten (Extra costs), Versicherung (Insurance), Gesamt (Total)
- Columns: Geschätzte Kosten, Geplante Kosten, Tatsächliche Kosten, Umsatz, Rabatt (editable %), Gewinn (Profit), Gesamt
- Each category has editable discount % and total override fields
- Category rows expandable (expand_more)
- Bottom summary: Rabatt (Discount) with total profit, Summe exkl. MwSt., MwSt. dropdown (Standard), Summe inkl. MwSt.
- Lock/unlock total (lock_open/lock)

**Zusätzliche Bedingungen (Additional Conditions):**
- Rich text editor (TinyMCE-style) with: Undo/Redo, Paragraph format, Font size, Bold/Italic/Underline/Strikethrough, Lists, Indent, Text color, Clear formatting, Insert link, Source code
- Template selector ("Vorlage anwenden" dropdown)
- Predefined conditions text area

**Bedingungen (Conditions):**
- Kaution (Deposit): currency input
- Rechnungszeitpunkt (Invoice moment): dropdown (e.g., "100% davor")

---

### 2.3 Lager / Warehouse (`#/warehouse`)

**Title:** Lager | Rentman Vermietungssoftware

**Header Controls:**
- Date navigation: chevron_left/right with current date display (e.g., "23 Marz 2026")
- Quick date buttons: "Heute" (Today), "Morgen" (Tomorrow)
- Location filter: "Lagerhaus Ost" dropdown
- Help link

**Toolbar:**
- Filter dropdown
- "Ansicht anpassen" (Customize view) dropdown
- Search
- Views settings dropdown
- **"Retour scannen" (Return scan)** button -- key warehouse workflow action

**Content:**
- **Fahrten (Trips)** section: expandable/collapsible
- **Geplante Mitarbeiter (Planned staff)** section: expandable/collapsible
- Timeline view showing projects and their warehouse status

**Sub-pages (from navigation):**
- Kombinationen (Combinations/Sets)
- Cross-Docking-Ubersicht (Cross-Docking Overview)
- Lager-Tracking-Log (Warehouse Tracking Log)

---

### 2.4 Personalplaner / Crew Planner (`#/planner`)

**Title:** Personalplaner | Rentman Vermietungssoftware

**Layout: Three-panel Gantt-style view**

**Left Panel - Employee List:**
- Dropdown: "Mitarbeiter" (can switch view)
- Search with tag filter
- Grid with checkboxes and expand_more for folders
- Employee categories organized in folders:
  - **Andere Freelancer**: Axel Keicher, Herbert Kowalski, Jochen Biersack, Marcel Gambrino, Mayara Dobinski, Michael Muller, Stefan Schackelgruber, Stefanie Kotzlowski
  - **Mitarbeiter Vollzeit**: Bruno Zampanowitsch, Gunther Hoffmann, John Dee, Thomas Schmidt
  - **Bevorzugte Freelancer**: Gunther Schabowski, Paule Kornsoffski, Peter Gravitronowitsch, Walter Rohrich
- Actions: watch_later, search, Einladen (Invite, disabled), Hinzufugen (Add, disabled)

**Center Panel - Gantt Timeline:**
- Filter bar with: filter_list, label (tags), location, views settings
- Toggle buttons: aspect_ratio (fullscreen), insights
- "Anzeigen" (Show) filter icons: group, local_shipping, group_add, mark_email_read, mail+access_time, drafts, flip_to_back
- **TreeGrid** structure with columns:
  - Checkboxes
  - Color indicators
  - Project names
  - Timeline bars (daily columns: Mo/Di/Mi/Do/Fr/Sa/So with dates)
- Projects listed with expandable sub-projects:
  - Schichten (Shifts)
  - Liveauftritt Band Soultrain
  - Gasthof Der Goldene Lowe Festinstallation
  - Musical 'Once upon a time'
  - EMF Light Show Dryhire
  - Abi Feier Dryhire
  - Philips Produktvorstellung
  - Soultrain Europatour (with sub-projects: Helsinki, Oslo)
  - Aufnahmetermin Fred Sheeran
  - Hendrik's Verkaufsprojekt
  - Sonstige Planung
  - Reservierungen
- **Timeline items** show crew/transport assignments with status:
  - `people (1/1) Stagehand` -- filled position
  - `people comment (0/1) Soundtechniker` -- unfilled with comment
  - `push_pin` -- pinned assignments
  - `local_shipping (0/1) Transport bis 6 m3` -- transport slots

**Summary row:** Shows totals per day (e.g., 1/2, 0/1, 1/16, 3/9, 9/11, 3/5, 10/12, 1/4)

**Right Panel - Employee Detail:**
- "Mitarbeiter" dropdown header
- Quick search
- Sort tabs:
  1. Geplant im Projekt (Planned in project)
  2. Verfugbar (Available)
  3. Ubereinstimmende Tags (Matching tags)
- Edit button
- Show filters: event, drafts, mark_email_read, mail+access_time, flip_to_back
- Availability grid: Verfugbarkeit column with daily timeline
- Print and open_in_new actions

**Header controls:**
- Date range: "Nachste Woche" dropdown with first_page/chevron_left/Heute/chevron_right/last_page navigation
- Zoom: remove/add buttons
- Map view button
- "Senden..." (Send) button with count badge ("13")

---

### 2.5 Mangel / Shortages (`#/subrent/shortages`)

**Title:** Mietengpässe | Rentman Vermietungssoftware

**Sub-pages:**
- **Mietengpässe** (Rental shortages)
- **Verkaufsengpässe** (Sales shortages)
- **Zumietungsjobs** (Sub-rental jobs)

**Features:**
- Standard list/grid view with filter bar
- Date range filter
- Search
- Custom filters
- Grid view settings
- Shortage detection system that compares planned vs available inventory

---

### 2.6 Finanzen / Finance

#### 2.6.1 Rechnungen / Invoices (`#/invoices`)

**Title:** Rechnungen | Rentman Vermietungssoftware

Standard list view with:
- Filter bar (date range, search, tags, custom filters)
- Grid with columns (configurable)
- Bulk action toolbar
- "Ansichten" (Views) customization
- "Externally invoiced" filter parameter

#### 2.6.2 Zu fakturieren / To Be Invoiced (`#/invoices/pending`)

Pending invoice management.

#### 2.6.3 Bestellungen / Purchase Orders (`#/invoices/purchase-orders`)

**Title:** Bestellungen | Rentman Vermietungssoftware

Purchase order tracking with:
- Purchase order status filter
- Standard grid view
- Links to sub-rental system

---

### 2.7 Materialien / Equipment (`#/equipment`)

**Title:** Materialien | Rentman Vermietungssoftware

#### 2.7.1 Materials List View

Standard grid with:
- Hierarchical folder structure (Audio, Show, etc.)
- Columns: Code, Name (in der Datenbank), sortable
- Item codes: Audio-348, Show-350, Audio-354, etc.
- Bulk actions on selection
- Views customization

**Sub-pages:**
- Seriennummern (Serial numbers)
- Lagerstandorte (Warehouse locations)
- Archivierte Materialien (Archived materials)
- Archivierte Standorte (Archived locations)

#### 2.7.2 Material Detail View (`#/equipment/{id}`)

**Example:** DJ Kit (Audio-348)

**Header:** Code + Name, Close/Save buttons

**Tab bar (11 tabs):**
1. **Daten** (Data)
2. **Seriennummern** (Serial Numbers)
3. **Einblicke** (Insights)
4. **Standardinhalt** (Default Content)
5. **Zubehör** (Accessories)
6. **Alternativen** (Alternatives)
7. **Lieferanten** (Suppliers)
8. **Webshop**
9. **Reparaturen** (Repairs)
10. **Bestandszählungen** (Stock Counts)
11. **History Log**

**Daten (Data) Tab - Left Column:**

**Material section:**
- Name (in der Datenbank): textbox
- Code: textbox (optional, with info tooltip)
- Ordner (Folder): dropdown (e.g., "Sets")
- In Planer wiedergeben (Show in planner): Yes/No dropdown
- Standard-Materialgruppe (Default material group): textbox with info tooltip
- Inhalt kann im Projekt bearbeitet werden (Content editable in project): Yes/No with info

**Eigenschaften (Properties) section:**
- Type badge: "Virtuelle Kombination - Vermietung"
- **Physisch/Virtuell (Physical/Virtual):** Radio buttons:
  - Physische Materialien (Physical materials)
  - Virtuelle Kombination (Virtual combination) -- checked
- **Vermietung/Verkauf (Rental/Sale):** Radio buttons (disabled when virtual):
  - Vermietung (Rental) -- checked
  - Verkauf (Sale)
- **Content capability:** Radio buttons:
  - Nein (No) - single material
  - Ja (Yes) - can have contents -- checked

**Finanzen (Finance) section:**
- Vermietpreis (Rental price): EUR 175.00 with info + cached/sync button
- Zumietungspreis (Sub-rental price): EUR 105.00 with info + cached
- Neupreis (Replacement price): EUR 0.00 with info + cached
- Break-even Preis: EUR 94.50 with info
- Rabattgruppe (Discount group): dropdown (e.g., "Rental")
- Faktorgruppe (Factor group): dropdown (e.g., "Default")
- MwSt Klasse (VAT class): dropdown (e.g., "Hoher Steuersatz")
- Hauptbuch - Haben (Ledger - Credit): dropdown (e.g., "Audio") with info, optional
- Hauptbuch - Soll (Ledger - Debit): dropdown (e.g., "Vermietung (soll)") with info, optional

**Extra Input Fields section (collapsible):**
- Custom field: Artikelbezeichnung in Englisch (Article name in English)

**Tasks/Notes/Files panel:**
- Same pattern as project: Aufgaben, Notizen, Dateien (1)

**Data Tab - Right Column:**

**Abbildung (Image):**
- Image display area
- "Mit Google suchen" (Search with Google) button
- "Klicke hier, um ein neues Bild auszuwählen" (Click to select new image) button

**Bestand (Stock):**
- Info note: "Stock of virtual combinations is calculated from individual materials in the combination"

**Kommentare (Comments):**
- Kommentar intern (Internal comment) with info
- Kommentar extern (External comment) with info
- Edit button ("Andern")

**Zusammensetzung (Composition):**
- Two tables:
  - **Inhalt (Contents):** Lists component materials with quantity and code
    - 2x Audio-345 - Pioneer CDJ2000 NXS2 Tabletop CD Player
    - 1x Audio-344 - Pioneer DJM900 NXS2 DJ Mixer
    - 1x Audio-346 - Shure SM58 Microphone
  - **Ist Teil von (Is part of):** Lists parent combinations
    - 1x Show-350 - Drive-in show without lighting
    - 1x Show-351 - Drive-in show with lighting
  - Each row has a launch button to navigate to the related item

**Tags:** With edit button

---

### 2.8 Kontakte / Contacts (`#/contacts/list`)

**Title:** Kontakte | Rentman Vermietungssoftware

Standard CRM list view with:
- Add contact button
- Filter bar (date range, location, search, tags, status, custom filter)
- Grid with configurable columns
- Folder organization
- Bulk edit actions
- Views customization

---

### 2.9 Mitarbeiter / Employees (`#/employees`)

**Title:** Mitarbeiter | Rentman Vermietungssoftware

Employee management with:
- Standard list/grid view
- Folder-based organization (Andere Freelancer, Mitarbeiter Vollzeit, Bevorzugte Freelancer)
- Employee profiles with skills, rates, availability
- Integration with Personalplaner

---

### 2.10 Fahrzeuge / Vehicles (`#/transport`)

**Title:** Fahrzeuge | Rentman Vermietungssoftware

Fleet management:
- Standard list/grid view
- Vehicle profiles (e.g., "LKW 1")
- Integration with transport planning in projects
- Capacity tracking (volume-based: "Transport bis 6 m3", "Transport bis 40 m3")

---

### 2.11 Aufgaben / Tasks (`#/tasks`)

**Title:** Aufgaben | Rentman Vermietungssoftware

Task management:
- Standard list/grid view
- Filter bar with date range, search, tags, custom filters
- Task states: Abgelaufen (Expired), completed, open
- Tasks linkable to projects, materials, or standalone
- Created by / date tracking
- Assignable to employees

---

### 2.12 Stundenerfassung / Time Tracking (`#/timeregistration/list`)

**Title:** Stundenerfassung | Rentman Vermietungssoftware

**Sub-pages:**
- **Stundenerfassung** (Time registration) -- main time entry list
- **Aktivitäten** (Activities) -- activity type management
- **Abwesenheitsanträge** (Leave requests)

Features:
- Default date filter: "Letzte Woche" (last week) for time entries
- Standard grid with filter bar
- Approval workflow implied
- Integration with crew planning

---

### 2.13 Werkstatt / Workshop

**Sub-pages:**
- **Reparaturen** (Repairs) -- repair tracking for equipment
- **Prüfungen** (Inspections) -- completed inspections
- **Zu prüfende Materialien** (Pending inspections) -- materials due for inspection
- **Verlorene Materialien** (Lost items) -- tracking lost equipment
- **Bestandszählungen** (Stock counts) -- inventory counting

---

### 2.14 Statistik / Statistics (`#/statistics`)

**Title:** Statistik | Rentman Vermietungssoftware

**Features:**
- "Statistik hinzufugen" (Add statistic) button
- Filter bar with search, tags, custom filters
- Grid of saved statistics with columns:
  - Typ (Type) -- sortable
  - Name -- sortable
- Statistics organized by type: Personal (Crew), etc.
- Views customization
- Bulk actions: Edit (Bearbeiten), more options

---

### 2.15 Kommunikation / Communication

**Sub-pages:**

#### Kommunikations-Log (`#/logscommunication`)
- Central log of all communication events
- Standard list view with filters

#### Gesendete E-Mails (`#/messages`)
- Sent email tracking
- Email templates and history

#### Erhaltene Notizen (`#/incomingnotes`)
- Received notes/communication
- Digital signing responses

---

## 3. Configuration System Detail

### 3.1 Configuration Panel Navigation

All config pages follow URL pattern: `#/configpanel/{section}`

The configuration panel has its own layout with sections organized in a sidebar or breadcrumb within the config area. Each page has:
- Heading with section name
- Info banner explaining the section
- Close/Save buttons
- Help/support links

### 3.2 Configuration Pages (Complete List)

| Config Page | German Title | Purpose |
|---|---|---|
| `time-and-location` | Zeit und Ort | Timezone, date format, location settings |
| `holidays` | Wichtige Tage | Holiday calendar, important dates |
| `numbers` | Nummernkreise | Number sequences for projects, invoices, etc. |
| `projecttypes` | Projekttypen | Project type definitions (Band, etc.) |
| `projecttemplates` | Projektvorlagen | Project templates |
| `inspections` | Regelmassige Prufungen | Periodic inspection schedules |
| `settings` | Stundenerfassung und Abwesenheit | Time tracking & leave settings |
| `statuses` | Lagerstatus | Warehouse status definitions |
| `customfields` | Extra Eingabefelder | Custom field definitions |
| `digitalsigning` | Digitale Unterschrift | Digital signature settings |
| `crewinvites` | Einladungen | Crew invite templates/settings |
| `writingpaper` | Briefpapier | Letterhead / document template |
| `onlinequotes` | Online-Angebote | Online quote portal settings |
| `title` | Anrede | Salutation/title options (Herr, Frau, etc.) |
| `general` | Finanzen | Financial defaults (see below) |
| `digital-invoicing` | Digitale Rechnungsstellung | Digital/e-invoicing |
| `quantitydiscountgroup` | Faktorgruppen | Quantity/factor discount groups |
| `globalrate` | Mitarbeitertarife | Global crew rates |
| `discountgroup` | Rabattgruppen | Discount group definitions |
| `invoicemoment` | Rechnungszeitpunkte | Invoice timing rules |
| `paymentcondition` | Zahlungskonditionen | Payment terms/conditions |
| `taxschemas` | MwSt Regelungen | VAT schemes |
| `taxclasses` | MwSt Klassen | VAT classes |
| `paymentmeans` | Zahlungsmittel | Payment methods |
| `ledger` | Hauptbucher | General ledger accounts |
| `conditions` | Zusätzliche Bedingungen | Additional terms & conditions |

### 3.3 Finance Configuration Detail (`#/configpanel/general`)

**Finanzielle Einstellungen (Financial Settings):**
- Wahrungszeichen (Currency symbol): textbox
- Standard Falligkeitsdatum fur Angebot/Vertrag (Default due date for quotes/contracts): days
- Standard-Ablaufdatum fur Mahnungen (Default expiry for reminders): days
- Standard Mehrwertsteuerregelung (Default VAT scheme): dropdown (e.g., "Standard")
- Standard MwSt Klasse (Default VAT class): dropdown (e.g., "Hoher Steuersatz")
- Standard MwSt.-Klasse Personalfunktionen (Default VAT class for crew): dropdown
- Standard MwSt.-Klasse Transportfunktionen (Default VAT class for transport): dropdown
- MwSt.-Klasse fur Versicherung (VAT class for insurance): dropdown (e.g., "Kein Steuersatz")
- MwSt.-Klasse der Zumiete in der Bestellung (VAT class for sub-rental in orders): dropdown

**Bankverbindung (Bank Details):**
- Bank (Name): textbox with info
- Kontonummer (Account number): textbox with info
- Bankleitzahl (BIC): textbox with info

**Geschaftsbedingungen fur Dokumente (T&C for Documents):**
- AGBs mit Angebot senden (Send T&C with quotes): Yes/No
- AGBs mit Vertrag senden (Send T&C with contracts): Yes/No
- AGBs mit Rechnung senden (Send T&C with invoices): Yes/No
- Allgemeine Geschaftsbedingungen (General T&C): File upload (drag & drop or click)

---

## 4. UX Patterns

### 4.1 Common UI Patterns

1. **List/Grid Pattern** -- Used everywhere (projects, materials, contacts, invoices, tasks, etc.)
   - Sortable column headers with sort direction indicators
   - Checkbox selection (single + select all)
   - Expandable folders/groups in grid
   - Row hover reveals inline action buttons
   - Single click selects, shows quick actions; double click opens detail

2. **Detail Panel / Inline Sidebar** -- When clicking a list item:
   - Right-side sliding panel (widget-sidebar) with lock_open/lock to pin
   - Shows summary info without navigating away
   - Close button to dismiss

3. **Full Detail Page** -- Double-click or explicit "Details" button:
   - Opens as a new tab in the tab bar
   - Tab-based sub-navigation within the entity
   - Close/Save buttons in header
   - Two-column layout (form fields left, contextual info right)

4. **Filter Bar** -- Consistent across all list views:
   - Date range dropdown with quick presets
   - Location/warehouse filter
   - Search toggle
   - View mode selector
   - Status filter (per entity type)
   - Tags filter
   - Custom filters
   - Active filters display

5. **Bulk Actions Toolbar** -- Appears when items selected:
   - Selection count display
   - Context-sensitive actions (Edit, Print, Timeline, etc.)
   - Close selection button
   - More options overflow menu

6. **Dropdown Pattern** -- Consistent `"Label arrow_drop_down"` pattern for all dropdowns

7. **Form Field Pattern:**
   - Label above field
   - Optional "(Optional)" indicator
   - Info tooltips (info icon)
   - Inline autocomplete lists
   - Currency prefix/suffix
   - Lock/unlock for calculated fields

8. **Expandable Sections** -- `expand_more`/`expand_less` toggles for collapsible content areas

9. **Tab Navigation** -- Used within detail pages for organizing sub-sections

10. **Gantt/Timeline View** -- Used in Personalplaner, Warehouse:
    - Left panel: entity list (employees, projects)
    - Right panel: time-based horizontal bars
    - Day/week navigation
    - Color-coded status bars

### 4.2 Icon System

Uses **Material Icons** (Google Material Design icon font):
- `event` -- Calendar
- `table_chart` -- Projects
- `category` -- Materials
- `account_circle` -- People/Crew
- `monetization_on` -- Finance
- `swap_horizontal_circle` -- Shortages/Exchange
- `local_shipping` -- Vehicles/Transport
- `assignment_turned_in` -- Tasks
- `watch_later` -- Time tracking
- `build` -- Workshop
- `poll` -- Statistics
- `dvr` -- Communication
- `settings` -- Configuration
- `dashboard` -- Dashboard
- `contact_phone` -- Contacts
- `search`, `filter_list`, `label`, `edit`, `delete`, `add`, `close`, `more_vert`, `print`, `lock`/`lock_open`, `launch`, `chevron_left`/`chevron_right`, `expand_more`/`expand_less`, `check`/`remove`, etc.

### 4.3 Color System

- Project status colors (customizable per project with hex color picker)
- Status badges with semantic colors (green for confirmed, red for cancelled, etc.)
- Material type indicators

### 4.4 Data Grid Features

- Column visibility and order customization ("Ansichten" settings)
- Sortable columns (ascending/descending with priority indicators)
- Column context menu (more_vert on column headers)
- Infinite scroll loading
- Row grouping by folder structure
- Accessibility: ARIA roles (grid, row, gridcell, columnheader, etc.)
- Keyboard support ("SPACE for context menu")

### 4.5 Document Generation

- "Projektdokument erstellen" / "Dokument erstellen" button available in projects and materials
- Print icon button
- Digital signing integration
- PDF generation for quotes, contracts, invoices
- Online quote portal
- Letterhead customization

---

## 5. Mobile / Responsive Patterns

Based on the application structure:
- The sidebar navigation is collapsible (chevron toggle), suggesting responsive adaptation
- The navigation has a `navigation__close_button` for mobile
- The layout uses percentage-based widths
- The crew planner has fullscreen toggles for panels
- No dedicated mobile breakpoint observed in this session, but the SPA architecture supports responsive behavior
- Zendesk chat widget (iframe) present in bottom-right

---

## 6. Key Takeaways for RentFlow

### 6.1 Core Data Model

The Rentman data model centers on:

1. **Projects** (the primary entity)
   - Have sub-projects
   - Link to contacts (client + venue)
   - Contain material lists, crew assignments, transport, additional costs
   - Track financial summary with category-level discounts
   - Support project types, templates, status workflow
   - Include document generation (quotes, contracts, invoices)
   - Have planning periods (calculated from sub-elements)

2. **Materials** (equipment)
   - Physical vs Virtual combinations
   - Rental vs Sale classification
   - Hierarchical composition (sets containing items, items belonging to sets)
   - Pricing: rental, sub-rental, replacement, break-even
   - Financial classification: discount groups, factor groups, VAT classes, ledger accounts
   - Serial number tracking
   - Stock/availability management
   - Supplier relationships

3. **Employees/Crew**
   - Organized in folders (full-time, freelancers, preferred freelancers)
   - Skills/functions (Techniker, Soundtechniker, Lichttechniker, Stagehand)
   - Rates (hourly, fixed price)
   - Availability tracking
   - Invitation/confirmation workflow

4. **Vehicles**
   - Capacity-based (volume: m3)
   - Linked to transport planning

5. **Contacts** (CRM)
   - Companies and contact persons
   - Addresses with Google Maps integration
   - Linked to projects as clients or venues

6. **Financial Entities**
   - Quotes (Angebote) with versions, status, views tracking
   - Contracts (Vertrage)
   - Invoices (Rechnungen) with payment tracking
   - Purchase orders (Bestellungen) for sub-rentals
   - Discount groups, factor groups, payment conditions, VAT schemas

### 6.2 Key Workflows

1. **Project Lifecycle:** Create → Set status (Option → Bestätigt → Gepackt → Am Veranstaltungsort) → Close
2. **Quote Workflow:** Create quote → Publish online → Track views → Digital signing → Convert to contract
3. **Equipment Planning:** Add materials to project → Check availability → Handle shortages (sub-rent or alternatives)
4. **Crew Planning:** Define functions needed → Invite/assign crew → Track acceptance → Manage availability
5. **Warehouse Workflow:** Pack items → Track status per project → Scan returns (Retour scannen) → Cross-docking
6. **Invoice Workflow:** To-be-invoiced items → Create invoice → Track payment → Handle reminders
7. **Workshop Workflow:** Report damage → Create repair → Track inspections → Manage lost items → Stock counts

### 6.3 Complexity Points (Where Rentman is Dense/Complex)

1. **Financial configuration** -- 15+ config pages for financial settings alone (VAT schemas, classes, ledgers, discount groups, factor groups, payment conditions, etc.)
2. **Crew planner** -- Three-panel Gantt view with many filter/display options and real-time slot filling
3. **Project detail** -- 11 tabs with deep sub-forms; the financial tab alone has a summary matrix, quote/contract management, invoice creation, rich text conditions editor
4. **Material properties** -- Physical/Virtual, Rental/Sale, composition hierarchies, multiple price fields, ledger assignments
5. **Shortage management** -- Cross-referencing availability across projects, locations, and time periods
6. **Document system** -- Quote/contract generation with digital signing, online portal, letterhead customization

### 6.4 UX Opportunities for RentFlow

1. **Simplify navigation** -- Rentman has ~35 sub-pages; consider progressive disclosure and smart defaults
2. **Reduce tab overload** -- Project detail has 11 tabs; consider a more compact layout with collapsible sections
3. **Improve financial overview** -- The finance tab is information-dense; consider visual dashboards over raw tables
4. **Modernize the Gantt view** -- Crew planner works but feels dense; opportunity for a cleaner timeline UX
5. **Better mobile support** -- Warehouse scanning (Retour scannen) is clearly a mobile workflow
6. **Streamline configuration** -- 25+ config pages could be organized into a wizard or grouped settings panel
7. **Inline editing** -- Rentman requires opening detail pages for most edits; consider more inline editing in grids
8. **Status visualization** -- Use color-coded progress indicators more prominently (Projektfortschritt pattern is good, expand it)
9. **Smart defaults** -- Many fields have complex dropdowns; AI-assisted defaults could reduce setup time
10. **Search & discovery** -- Global search exists but could be enhanced with command-palette style (like Rentman's `/ ` shortcut, but richer)

### 6.5 Technical Architecture Notes

- **SPA (Single Page Application)** with hash-based routing
- **Accessibility:** Uses ARIA roles extensively (grid, treegrid, row, gridcell, columnheader, button, checkbox, radio, textbox, combobox, etc.)
- **Keyboard support:** Documents keyboard shortcuts (alt+e for edit, alt+/ for quick lookup, / for search)
- **Data grid:** Custom grid implementation with virtual scrolling, sortable/reorderable columns
- **Rich text:** TinyMCE-style editor for conditions/notes
- **Maps:** Google Maps integration for location display and distance calculation
- **File handling:** Drag-and-drop upload, PDF generation, image search via Google
- **Real-time:** Notification system (notifications_none bell icon)
- **Third-party:** Zendesk chat widget, Google Maps API
- **Localization:** Fully localized to German (all UI strings, date formats)

---

## Appendix: Snapshot Files

All accessibility snapshots saved to: `C:\Users\jecke\rentflow\docs\analysis\snapshots\`

| File | Section |
|---|---|
| `01-dashboard.md` | Dashboard |
| `02-projects-list.md` | Projects list |
| `02b-projects-list-scrolled.md` | Projects list with full navigation |
| `03-project-material-tab.md` | Project material tab |
| `04-warehouse.md` | Warehouse |
| `05-shortages.md` | Shortages (Mietengpasse) |
| `06-materials-list.md` | Materials list |
| `07-invoices.md` | Invoices |
| `08-contacts.md` | Contacts |
| `09-tasks.md` | Tasks |
| `10-timetracking.md` | Time tracking |
| `11-communication.md` | Communication log |
| `12-statistics.md` | Statistics |
| `13-material-detail.md` | Material detail (DJ Kit) |
| `config-01-time-and-location.md` | Config: Time & Location |
| `config-08-customfields.md` | Config: Custom Fields |
