# Portfolio Dashboard: Design

Date: 2026-10-29 · Owner: Chirag · Status: draft, pending approval

## 1. Intent

Groww shows P&L only against the lot currently held. Example: buy an ETF for ₹1,00,000, sell for ₹88,000 (−₹12,000), later re-buy for ₹88,000, now worth ₹90,000. Groww shows **+₹2,000**, but the lifetime figure for that ETF is **−₹10,000**, before charges.

This app shows the **lifetime** picture. For every stock, ETF and mutual fund ever held, including ones fully exited, it shows what was invested, what came back, what is still held, and the realized/unrealized split. These roll up into a Stocks dashboard, an MF dashboard, and one top-level Overview.

## 2. Goals

- Per-instrument lifetime view for every instrument ever held, current quantity 0 included.
- Stocks dashboard and Mutual Funds dashboard, each with totals.
- Overview combining both.
- Every level shows: **current invested, current value, overall P&L, unrealized, realized, dividends, charges paid**.
- Graphs, Groww-style dark UI (approved mockup: `mockups/dashboard-v1.html`).
- Current prices refresh automatically; transactions update on demand via file upload.
- Completely free to run.

## 3. Non-goals (for now)

- Bonds, FDs, gold, F&O, US stocks, or anything other than Indian stocks/ETFs and mutual funds.
- A light theme.
- Hosting, multi-user support, or a login screen. It runs locally on one laptop first; hosting comes later.
- Real-time tick streaming.
- Dividend data. The dividends line exists but is hardcoded to ₹0; its data source will be decided later.
- Tax reports or capital-gains computation for ITR.
- Placing orders.

## 4. Anti-goals (failure even if it "works")

- **Numbers that can't be trusted.** If the lifetime P&L for an instrument is wrong, the app has failed. When data is incomplete or inconsistent, the app must flag it visibly, not silently show a number.
- **Anything not above-board.** No scraping Groww, no use of Groww's private/internal web APIs, no storing Groww credentials, no browser extensions that piggyback on the Groww session. (User decision.)
- **Anything that costs money.** No paid APIs, including the Groww Trade API, and no paid hosting.

## 5. Constraints and findings that shaped this

The Groww official Trade API was checked in Oct 2026 and is not usable here:
- It's paid (~₹499/mo + GST).
- It has no OAuth or "Login with Groww" for third-party apps. Its access token expires daily at 6 AM and needs daily manual approval.
- Its order list returns **today's orders only**, with no date range and no charges fields.
- It doesn't support mutual funds.

So the data comes from **files the user downloads and uploads manually** (user decision):
- **Stocks/ETFs:** Groww's downloadable reports (Profile → Reports), issued per financial year as Excel.
- **Mutual funds:** the CAS (Consolidated Account Statement) from CAMS/KFintech (a password-protected PDF) **and** Groww's MF reports, merged and deduplicated.

All parsed data is stored in the app's own database, which is the source of truth from then on.

## 6. Decisions

### Decided by the user
| Topic | Decision |
|---|---|
| Data acquisition | Manual upload only. One-time backfill of all years, then "Update data" re-uploads the current FY's files. |
| MF sources | Both CAS and Groww MF reports, parsed and stored in our own format. |
| Freshness | T-1-ish transactions (as of last upload) are acceptable. |
| Overall P&L formula | **Overall = Realized + Unrealized + Dividends − Charges** |
| Dividends | Shown as a separate line and included in overall P&L. ₹0 for now. |
| Where it runs | Locally on the laptop first; hosting later. |
| Stack | Next.js frontend, Go backend. |
| Look | Groww-like, dark only. Mockup `mockups/dashboard-v1.html` approved. |
| Project location | `~/chirag/portfolio` |
| Data-source architecture | App-wide, pluggable **data sources**. Groww is one source, CAS another. More (other brokers, NSDL/CDSL CAS, …) must be addable later for stocks as well as MFs, without reworking the ledger. |
| Source filter | An app-wide filter by data source (e.g., Groww, CAS). No filter selected = everything. |
| MF scope | Include **all** funds in the CAS, not just Groww-bought ones. |
| Stock history completeness | For now, assume every stock/ETF came from a Groww buy. Opening balances and transfers-in come later with additional sources. |
| Corporate actions (v1) | Detect inconsistencies → warn → user enters the adjustment via a simple form (split/bonus ratio, merger/demerger mapping, rename + date). An automatic corporate-actions feed comes later. |

### My calls (stated during the conversation, not objected to)
| Topic | Call |
|---|---|
| Database | SQLite, a single file. Swap to Postgres when hosting. |
| Realized/unrealized split | **FIFO** per instrument, matching Indian tax treatment. The lifetime total doesn't depend on the method; only the split does. |
| Excel parsing | Go, `excelize`. |
| CAS PDF parsing | The Python `casparser` library, invoked by the Go backend as a subprocess (small CLI script that emits JSON). Go has no mature equivalent. |
| Stock prices | Yahoo Finance (free, unofficial, `SYMBOL.NS`), polled about every 60s during NSE hours (09:15–15:30 IST, weekdays). |
| MF NAVs | AMFI's official daily NAV file (free), fetched once after market close. |
| Auth | None while local. The backend binds to 127.0.0.1 only. |

## 7. Data and calculations

### 7.0 Data sources (app-wide)
- A **data source** is a pluggable module: it declares which instrument types it covers, which file types it accepts, and a parser that emits canonical transactions (§7.2). v1 sources: **Groww** (stock xlsx, MF xlsx) and **CAS** (CAMS/KFintech PDF).
- Adding a source = adding a module. The ledger, rollups and UI don't change.
- Every stored transaction records the source(s) it was seen in. When two sources describe the same transaction (e.g., CAS + Groww for an MF purchase), it is stored once and tagged with both.
- **Source filter (my call on semantics):** selecting sources keeps the transactions tagged with any selected source, and all numbers are recomputed from that subset. No selection = all transactions. The filter is global (nav bar) and applies to every screen.

### 7.1 Instruments
- Identify instruments by **ISIN** wherever available. Keep the exchange symbol (stocks) and scheme name/AMFI code (MFs) as attributes.
- Instrument types: `STOCK` (ETFs are included here and show on the Stocks dashboard, matching Groww), `MF`.

### 7.2 Transactions (canonical form)
Each row: instrument, date (and time if available), type (BUY/SELL for stocks; PURCHASE/REDEMPTION/SWITCH_IN/SWITCH_OUT for MFs; plus ADJUSTMENT for user-entered corporate actions), quantity/units, price/NAV, gross amount, charges breakdown (brokerage, STT, exchange txn, SEBI fee, stamp duty, GST, DP charges, other), source file, and source row reference.

### 7.3 Per-instrument numbers
- **Current quantity** = Σ buys − Σ sells.
- **Invested now** = FIFO cost of the lots still held.
- **Current value** = current quantity × latest price/NAV.
- **Unrealized** = current value − invested now.
- **Realized** = Σ over sells of (sell amount − FIFO cost of the units sold).
- **Charges** = Σ all charges on all transactions. These are **kept out of cost basis**, so they appear once, in their own line.
- **Dividends** = 0 (placeholder).
- **Overall P&L** = realized + unrealized + dividends − charges.
- **Total ever invested** = Σ buy amounts (used for the lifetime % on the Overview).

Check against the ETF example: realized −12,000, unrealized +2,000, charges 186 → overall −10,186. This must be a unit test.

### 7.3a Corporate actions (v1)
- Detection: a sell exceeding held quantity, an instrument whose price feed stops (rename/delisting), or a holdings mismatch vs. a statement → warning on that instrument.
- Fix: a form where the user enters the action. Split/bonus (ratio + date) rebases lot quantities and per-unit cost, with total cost unchanged. Merger/demerger (from ISIN → to ISIN(s), ratio, cost split %, date) moves lots. Rename maps old symbol/ISIN to new so prices keep working.
- These entries are stored as a "Manual" data source, so they're visible and undoable.

### 7.4 Rollups
Stocks dashboard = Σ over stock instruments. MF dashboard = Σ over MF instruments. Overview = Stocks + MF. Every figure is a straight sum.

### 7.5 Value-over-time chart (Overview)
Rebuild daily holdings from transactions, multiply by historical daily close/NAV, and plot "Current value" against "Invested". **My call:** "Invested" on each day = FIFO cost of the lots held that day, replayed day by day, so the line's last point equals the "Currently invested" tile. Historical prices come from Yahoo Finance for stocks and from mfapi.in (free) for MF NAV history. Cache prices in SQLite so the chart doesn't refetch.

## 8. Import flow

1. The **Update data** page accepts Groww stock reports (.xlsx), Groww MF reports (.xlsx) and CAS (.pdf, which prompts for its password; the password is used once and never stored).
2. The backend parses each file into canonical transactions and computes a **content fingerprint** per transaction (instrument + date + type + qty + price + amount, plus order/trade ID when present).
3. Fingerprints already present are skipped, so re-uploading the same or an overlapping FY file is safe (idempotent).
4. MF merge: CAS and Groww MF records describing the same transaction are matched (ISIN/scheme + date + units within tolerance + amount) and stored once. Each row keeps a link to both sources.
5. The import summary shows rows added, skipped as duplicates, and **warnings**. Examples: a sell exceeding held quantity; an unknown instrument; computed holdings disagreeing with CAS closing balance.
6. Instruments with warnings show a warning badge wherever they appear, and their numbers are marked "may be incomplete". (This is the anti-goal guard.)
7. Uploaded raw files are kept in `data/uploads/` (gitignored) so the DB can be rebuilt by re-parsing.

## 9. Prices and freshness

- **Stocks:** during market hours the backend polls Yahoo for held stocks about every 60s. Outside market hours it uses the last close.
- **MFs:** the AMFI NAV file is fetched daily after about 21:00 IST, plus on demand.
- The **Refresh** action on the nav bar re-pulls prices immediately. **Update data** is the upload flow.
- The nav bar always shows two timestamps: "Prices updated …" and "Transactions as of …" (the latest transaction date in the DB).
- The frontend re-queries the backend periodically (about every 60s) while open. No websockets are needed.

## 10. Screens (see `mockups/dashboard-v1.html`)

1. **Overview:** current value, currently invested; overall lifetime P&L (₹ and % of total ever invested); tiles for Unrealized, Realized, Dividends, Charges; value-vs-invested chart with 1M/6M/1Y/3Y/All; Stocks and MF summary cards linking to their dashboards.
2. **Stocks dashboard:** tiles (current value, overall P&L, unrealized, realized, charges); lifetime-P&L-by-instrument diverging bar chart; table of all instruments ever held with an All/Holding/Exited filter. Columns: name + qty, invested now, current value, unrealized, realized, charges, overall P&L. Sortable.
3. **Mutual Funds dashboard:** same layout as Stocks.
4. **Instrument detail:** name, symbol, LTP; overall lifetime P&L headline; a callout showing what Groww would display whenever it differs from the lifetime figure. **My call:** compute "Groww's number" the way Groww does: weighted-average buy price of the current holding (the average resets after a full exit), so unrealized = qty × (LTP − avg price). This number appears only in the callout; all other figures use FIFO; tiles; a price chart with buy/sell markers; the full transaction list with per-sell realized P&L.
5. **Update data:** the upload page from §8 (pick the data source, then upload).
6. **Corporate action form:** opened from an instrument's warning badge (§7.3a).

The nav bar on every screen carries the global **data-source filter** (§7.0).

Theme: near-black background (#121212), cards (#1b1b1b), Groww green (#00b386) for gains, orange-red (#eb5b3c) for losses, Inter font. Dark only. Indian number formatting (₹1,00,000).

## 11. Architecture

```
Next.js (localhost:3000)  ──HTTP/JSON──▶  Go API (127.0.0.1:8080)
                                              ├─ sources/: pluggable data sources: groww (xlsx via excelize), cas (python casparser subprocess), manual
                                              ├─ ledger: FIFO engine, rollups
                                              ├─ prices: Yahoo poller, AMFI NAV, mfapi history
                                              └─ SQLite (data/portfolio.db)
```

- One command starts everything (e.g., `make dev`). Prerequisites: Go, Node, Python 3 with `casparser` installed in a project-local venv.
- `data/` is gitignored (DB, uploads).

## 12. What's needed from the user

- **Sample files, provided during implementation** (the user will download them then; sequence parser work after the ledger/UI, using fixture data until then): at least one Groww stock report per type available (order history / P&L / tax report: whichever contain transaction-level rows and charges), one Groww MF report, and one CAS PDF. The exact Groww column layouts are unverified, and the parsers must be built against real files.
- Their PAN-based CAS password at upload time (not stored).
- No API keys are needed.

## 13. Left to the builder

- Choice of Next.js charting library, as long as it can match the mockup.
- Go HTTP router, migrations tool, project layout.
- Exact SQLite schema, consistent with §7.
- Which Groww report is primary for stock transactions once the sample files are seen. Prefer the one with per-trade charges.
- How to allocate charges reported only in aggregate (e.g., per contract note): spread them pro-rata by trade value within that note.
- If the Groww Excel reports turn out to lack per-trade charges: **my call** is to compute them from the standard published rates (brokerage, STT, exchange, SEBI, stamp duty, GST) and label them "estimated" in the UI. A contract-note PDF parser can be added later as another data source.

## 14. Open items for later

- Dividend data source.
- Automatic corporate-actions feed (v1 is manual entry, §7.3a).
- More data sources: other brokers, NSDL/CDSL CAS, opening balances / transfers-in.
- Hosting and auth.
