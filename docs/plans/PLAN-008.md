# PLAN-008: Web MVP Vue untuk data pribadi

Status: Siap diimplementasikan — refinement 17 September 2026
Issue: [ISSUE-008](../issues/ISSUE-008-web-mvp.md)
Repositori implementasi: `monelog-app`
Kontrak API: [`monelog-api/api/openapi.yaml`](../../api/openapi.yaml), commit `bfe9f0398d457af963c5eb17d57db48df13129af` atau revisi yang kompatibel
Persyaratan: FR-01, FR-02, FR-03, FR-04, FR-05, FR-06, FR-07, FR-11, FR-17

## Keputusan refinement

Implementasi ini adalah browser MVP untuk data pengguna yang sedang login. Semua request aplikasi memakai hanya route personal `/api/v1/...`; UI tidak menerima `user_id`, tidak membangun route `/api/v1/admin/...`, dan tidak menyediakan layar pengelolaan admin. `GET /me` tetap dimuat untuk identitas, timezone, currency, dan perubahan session; nilai `role` hanya informasi profil, bukan keputusan keamanan di client. Kepemilikan, role, akun aktif, dan otorisasi tetap diperiksa backend pada setiap request.

Rujukan visual adalah layar Android Home, detail hari, tambah transaksi, dan laporan: pertahankan alur ringkasan hari, daftar transaksi, form tanggal/tipe/kategori/jumlah/judul, serta laporan numerik. Tangkapan layar tidak ada sebagai berkas di dua workspace saat refinement ini dibuat, sehingga warna, ikon, dan jarak piksel tidak boleh dianggap terkonfirmasi sampai aset referensinya tersedia. Dokumen `monelog-app/docs/api-integration.md` adalah API SQLite lama (`/api`, USD, angka JSON) dan tidak dipakai; frontend mengikuti OpenAPI Monelog v1 yang memakai IDR dan string uang.

MVP tetap online-only. `isDelete=false` adalah data aktif dan `isDelete=true` adalah Trash; delete tidak berarti hard delete. Tambahkan Trash sebagai bagian dari alur kategori/transaksi agar restore yang diwajibkan FR-03 dapat dilakukan, bukan sebagai fitur admin.

## Prasyarat dan dependensi tepat

Scaffold baru diperlukan karena `monelog-app` saat ini hanya berisi README dan tiga dokumen, tanpa `package.json`, `src`, atau konfigurasi frontend. Gunakan Node `>=22.22.2`; versi itu memenuhi engine Vitest dan jsdom yang dipatok di bawah.

`package.json` harus memakai script berikut:

```json
{
  "scripts": {
    "dev": "vite",
    "build": "vite build",
    "preview": "vite preview",
    "test": "vitest run",
    "test:watch": "vitest",
    "test:coverage": "vitest run --coverage",
    "test:e2e": "playwright test"
  },
  "dependencies": {
    "pinia": "4.0.3",
    "vue": "3.5.43",
    "vue-router": "5.3.1"
  },
  "devDependencies": {
    "@playwright/test": "1.63.0",
    "@vitejs/plugin-vue": "6.0.9",
    "@vitest/coverage-v8": "5.0.1",
    "@vue/test-utils": "2.5.1",
    "jsdom": "30.1.0",
    "vite": "8.3.0",
    "vitest": "5.0.1"
  }
}
```

Tidak tambahkan axios, pustaka tanggal, pustaka decimal, UI kit, icon pack, chart library, TypeScript, Capacitor, atau state/query cache. Browser `fetch`, `crypto.randomUUID`, `Intl`, Vue, Router, Pinia, CSS biasa, dan utilitas string lokal sudah cukup. Commit `package-lock.json` bersama manifest agar versi benar-benar terkunci.

Vite menggunakan `VITE_API_BASE_URL=/api/v1` secara default. Konfigurasi development mem-proxy `/api` ke URL API lokal yang dikonfigurasi agar browser tetap pada origin Vite; backend development harus memasukkan `http://localhost:5173` dalam `AUTH_ALLOWED_ORIGINS`. Production harus menyajikan UI dan `/api` pada origin HTTPS yang sama atau menyediakan origin yang terdaftar secara eksplisit. Login, refresh, dan logout browser wajib memakai `credentials: 'include'`; backend mengirim cookie refresh `__Host-monelog-refresh` HttpOnly dan cookie CSRF `__Host-monelog-csrf` yang dapat dibaca JavaScript. Setiap refresh/logout mengirim body `{"client_type":"web"}` dan header `X-CSRF-Token` dari cookie CSRF. Access token tidak pernah masuk cookie aplikasi, localStorage, sessionStorage, IndexedDB, URL, atau log.

## File dan folder yang akan dibuat

```text
monelog-app/
├── .env.example                         # VITE_API_BASE_URL dan target proxy non-rahasia
├── .gitignore                            # node_modules, dist, .env.local, test artifacts
├── index.html
├── package.json
├── package-lock.json
├── vite.config.js                        # Vue plugin dan proxy /api untuk development
├── vitest.config.js                      # jsdom, setup, coverage V8
├── playwright.config.js                  # Chromium, base URL, screenshot/trace saat gagal
├── src/
│   ├── main.js                           # Pinia, router, bootstrap auth, mount
│   ├── App.vue                           # RouterView, global live region dan session banner
│   ├── router/index.js                   # Route records dan auth guard
│   ├── services/
│   │   ├── http.js                       # fetch envelope, error normalisasi, 401 refresh sekali
│   │   ├── auth-api.js                   # login, refresh, logout, me, update profile
│   │   ├── categories-api.js             # personal category/Trash lifecycle endpoints
│   │   ├── transactions-api.js           # personal transaction, detail, daily, Trash endpoints
│   │   └── reports-api.js                # personal summary dan breakdown endpoints
│   ├── stores/
│   │   ├── auth.js                       # token memori, me, bootstrap, logout, session expiry
│   │   ├── categories.js                 # active/Trash by type and lifecycle invalidation
│   │   ├── transactions.js               # list/detail/day pages, cursors, mutations
│   │   ├── daily.js                      # daily summaries and current-day zero fallback
│   │   └── reports.js                    # range, summary, breakdown, report transaction page
│   ├── utils/
│   │   ├── money.js                      # validate/normalise string API and IDR display
│   │   ├── dates.js                      # YYYY-MM-DD, timezone today, Indonesian date display
│   │   ├── errors.js                     # safe Indonesian status/error messages
│   │   └── request.js                    # AbortController and request generation helpers
│   ├── components/
│   │   ├── AppShell.vue                  # desktop sidebar/mobile bottom navigation
│   │   ├── AppHeader.vue
│   │   ├── BottomNav.vue
│   │   ├── PageState.vue                 # loading, empty, retry, network/session messages
│   │   ├── ConfirmDialog.vue              # native-dialog based destructive confirmation
│   │   ├── MoneyText.vue                  # string decimal to formatted IDR only
│   │   ├── DateText.vue
│   │   ├── DailySummaryCard.vue
│   │   ├── DailySummaryList.vue
│   │   ├── TransactionList.vue
│   │   ├── TransactionRow.vue
│   │   ├── TransactionForm.vue            # shared add/edit form; owns draft and validation
│   │   ├── CategoryList.vue
│   │   ├── CategoryForm.vue
│   │   ├── TrashPanel.vue                 # transaction/category tabs and restore
│   │   ├── ReportTotals.vue
│   │   └── ReportBreakdownTable.vue
│   ├── views/
│   │   ├── LoginView.vue
│   │   ├── HomeView.vue
│   │   ├── DayDetailView.vue
│   │   ├── TransactionCreateView.vue
│   │   ├── TransactionEditView.vue
│   │   ├── CategoriesView.vue
│   │   ├── SettingsView.vue
│   │   ├── ReportsView.vue
│   │   ├── TrashView.vue
│   │   └── NotFoundView.vue
│   └── styles/
│       ├── tokens.css
│       ├── base.css
│       ├── layout.css
│       └── components.css
└── tests/
    ├── setup.js
    ├── unit/{money,dates,http,router}.spec.js
    ├── components/{TransactionForm,ConfirmDialog,PageState,ReportBreakdownTable}.spec.js
    ├── stores/{auth,categories,transactions,daily,reports}.spec.js
    └── e2e/{auth,transactions,categories,reports,responsive}.spec.js
```

Files only become implementation changes in `monelog-app`; this plan remains canonical in `monelog-api/docs/plans`. Existing `.idea/` is an untracked local IDE directory and remains untouched.

## Komponen, layar, dan responsivitas

`AppShell` owns application navigation and no financial data. At 360 px it uses a single-column page, tap targets at least 44 px, a fixed bottom navigation for **Beranda**, **Laporan**, and **Pengaturan**, and a primary **Tambah** action without horizontal scrolling. At desktop (`min-width: 768px`) it moves navigation to a visible sidebar, keeps readable line lengths, and uses multi-column summary/cards only where space permits. Ordinary CSS custom properties provide colors, spacing, typography, focus ring, success/danger tones, and reduced-motion behavior. Text labels and signed amount text always accompany color.

`LoginView` has labelled email/password controls, password visibility toggle, submit status, focused validation/error summary, and the Indonesian messages **Masuk**, **Email**, **Kata sandi**, **Coba lagi**, and **Sesi Anda telah berakhir**. It posts `{email, password, client_type: "web"}`. Successful login stores only `access_token` in the auth Pinia store, then obtains `/me` before routing to the intended page or Beranda. Failed login keeps credentials in the DOM and shows a non-sensitive error; it never exposes whether the account exists.

`HomeView` obtains the profile-timezone current date, requests an initial 30-day `/daily-summaries` page, renders today’s income/expense/difference with zeroes when today has no row, then shows earlier dated rows with **Muat lagi** from `page.next_cursor`. A day row routes to `DayDetailView`; it does not infer totals in JavaScript.

`DayDetailView` receives a `YYYY-MM-DD` route param, displays `DateText`, fetches that day’s `/daily-summaries` plus active `/transactions?start_date=<date>&end_date=<date>`, and renders totals plus a cursor-paginated `TransactionList`. Each item opens its edit route. The add action preserves the selected date in the create route query.

`TransactionForm` is used by `TransactionCreateView` and `TransactionEditView`; views own route loading/navigation while the component owns the editable draft, client validation, and focus. Fields are **Tanggal**, **Tipe**, **Kategori**, **Jumlah**, and **Judul**. Type toggles switch the active category list and clear an incompatible selection. Amount is a positive, ungrouped input (`inputmode="decimal"`): accept `1234,50` or `1234.50`, reject grouping separators/negative/zero/more than two fractional places, normalise to a two-place API string, and send no numeric JSON values. The display-only `MoneyText` groups the integer string with `.` and uses `,` for cents, e.g. `Rp1.000.000,00`, without binary floating-point arithmetic. `DateText` formats `YYYY-MM-DD` through `Intl.DateTimeFormat('id-ID', { timeZone: 'UTC' })`; profile timezone determines today and the form’s future-date maximum.

Create sends `transaction_date`, `type`, `category_id`, exact `amount`, trimmed `title`, and `client_request_id: crypto.randomUUID()`. The form retains one request ID for an unchanged retry after a network failure; any later draft change creates a new ID before submission, preventing an idempotency conflict for a different payload. Edit loads only the personal active detail endpoint and PATCHes the contract fields plus current `version`. Delete opens `ConfirmDialog`, sends `If-Match: "<version>"`, and returns to its day after 204. Because DELETE increments version without returning it, `TrashPanel` reloads its Trash detail before sending restore `{version}`. All 4xx/5xx/network failures retain every draft value, retain focus context, re-enable submit, and present an Indonesian form-level message; local field validation links errors with `aria-describedby` and focuses the first invalid field.

`CategoriesView` has **Pemasukan** and **Pengeluaran** tabs, each with active list, create/edit form, delete confirmation, and a link to **Sampah**. `TrashView` uses `TrashPanel` for `isDelete=true` transaction/category pages and restores with the freshly loaded version. Deleted categories remain visible as historical `category_name` on transactions but never appear in add/edit selectors. Categories store only its own active/deleted data, invalidates type lists after a mutation, and never guesses a version.

`SettingsView` loads `/me`, presents the signed-in email, IDR currency, profile timezone select/input, version-aware `PATCH /me`, a link to Sampah, and logout confirmation/action. It does not expose role administration or account-management UI. `ReportsView` provides **7 hari terakhir**, **30 hari terakhir**, and **Kustom**. Custom requires both inclusive dates and `start_date <= end_date`; presets send only `range`. It requests `/reports/summary` and `/reports/breakdown` together, displays numeric income/expense/difference, top income/expense categories, and an accessible table for **Minggu**, **Bulan**, or **Kategori**. A numeric transaction table for the resolved response `period.start_date`/`end_date` uses `/transactions`; there is no graph UI, export, or client-derived financial total.

All pages use explicit loading skeleton/status, empty copy/action, retryable network error, validation/conflict error, and session-expired state. The app applies `aria-busy` to loading regions, `aria-live="polite"` to status feedback, `aria-live="assertive"` for submit failures, visible `:focus-visible`, semantic headings/list/table markup, real `label`/`for` pairs, keyboard-operable native buttons/links/dialog actions, Escape for dialogs, and focus return to the invoking action. Page changes move focus to the page heading.

## API client, Pinia, dan route guards

`services/http.js` is the only code that calls `fetch`. It serializes JSON, parses `{data}` and `{error:{code,message}}`, represents status/code/request failure in a normalised error, uses `credentials: 'include'`, adds `Authorization: Bearer <in-memory token>` only to protected calls, and never persists responses. For a protected 401 it awaits one shared refresh operation and retries that original request once. It does not retry non-idempotent mutations automatically. Concurrent 401 responses share the same refresh promise; failed refresh/logout clears all user stores, aborts in-flight data requests, and moves to Login with `?alasan=sesi-berakhir`.

`auth-api.js` owns the unauthenticated login/refresh/logout transport. Its refresh/logout reads only the non-HttpOnly CSRF cookie and sends it in `X-CSRF-Token`; the refresh secret itself is neither read nor represented by JavaScript. `auth.js` exposes `bootstrap`, `login`, `logout`, `refreshAccessToken`, `loadMe`, and `clearSession`. Bootstrap tries web refresh once after reload, then `/me`; no token remains after a hard refresh if the server cookie flow fails. A session is initialized before protected data stores run and all stores reset on logout/session expiry.

The remaining API modules expose only personal endpoints and contain no route/view logic. `categories.js` caches lists keyed by `{type,isDelete}`. `transactions.js` caches only the current page/detail key `{date range, filters, isDelete, cursor}` and uses a generation number plus `AbortController` to discard stale responses on route/session changes. `daily.js` owns daily summary pagination and zero fallback. `reports.js` owns a validated range/filter and the matching summary, breakdown, and transaction page; no store calculates totals by parsing floats. After create/update/delete/restore, the affected stores invalidate/refetch the active day, day detail, daily history, reports, and category selector only as needed.

Routes are exactly:

| Path | Name | Guard / purpose |
| --- | --- | --- |
| `/login` | `login` | public-only; signed-in users return to Beranda |
| `/` | `home` | requires authenticated session |
| `/hari/:date` | `day-detail` | requires auth; validate `YYYY-MM-DD` before fetch |
| `/transaksi/tambah` | `transaction-create` | requires auth; optional valid `?tanggal=` |
| `/transaksi/:id/edit` | `transaction-edit` | requires auth; UUID is opaque and backend remains authority |
| `/kategori` | `categories` | requires auth |
| `/sampah` | `trash` | requires auth |
| `/laporan` | `reports` | requires auth |
| `/pengaturan` | `settings` | requires auth |
| `/:pathMatch(.*)*` | `not-found` | authenticated shell or Login redirect, as appropriate |

The global `beforeEach` waits for one `auth.bootstrap()` result. `requiresAuth` without a valid `/me` redirects to `/login?lanjut=<encoded path>`. `publicOnly` redirects an authenticated session to the sanitized `lanjut` target or `/`. Navigation guards do not authorize resource IDs or roles and do not call admin endpoints; backend 401/403/404 remains the source of truth.

## Urutan implementasi

1. Record Node version, create the Vite JavaScript scaffold, exact manifest/lockfile, environment example, Vite proxy, Vitest/Playwright configuration, CSS tokens/base, and build a blank app. Install Playwright Chromium explicitly with `npx playwright install chromium`.
2. Implement string money/date/error/request utilities and their unit tests before any form or financial rendering. Build `http.js`, auth API/store, bootstrap, global session handling, and route guards. Verify login, refresh after browser reload, logout, and expired session against the backend contract.
3. Build reusable app shell, status component, confirmation dialog, money/date display, focus management, and mobile/desktop CSS. Then implement Login, Settings, and Home/daily store/view.
4. Implement personal transaction API/store, day detail, shared add/edit form, idempotent create handling, versioned edit/delete, and route/query return behavior. Invalidate/refetch daily and report data after confirmed mutations.
5. Implement category API/store/view and Trash/restore for both resource types. Ensure the delete-then-reload-trash version sequence is used and deleted categories are excluded from form options.
6. Implement report API/store/view: preset/custom validation, concurrent numeric summary/breakdown loading, group table, and transaction table for the resolved report period. Do not add chart or export code.
7. Add focused component/store tests, then Playwright user flows against a controlled API fixture. Test keyboard and 360/desktop layouts, build production assets, review the diff, and capture actual command results in the issue before status changes.

## Tests dan bukti

Unit tests (`tests/unit`) cover:

- `money.spec.js`: valid API money boundaries, comma/dot draft normalisation, rejection of grouping/negative/zero/three decimals, signed difference display, and `Rp` grouping without `Number`/float conversion.
- `dates.spec.js`: calendar date validation, Indonesian display without off-by-one timezone shifts, `Asia/Jakarta` current date, and report inclusive-range validation.
- `http.spec.js`: bearer header only in memory, 401 → one shared refresh → retry once, refresh failure clears session, generic 4xx/5xx/network error mapping, and no automatic POST replay.
- `router.spec.js`: bootstrap redirect, protected intended-location redirect, public-only redirect, invalid local date/query handling, and no role/admin route rule.

Component tests (`tests/components`) use Vue Test Utils and jsdom:

- labelled controls, visible focus/error linkage, keyboard confirm/cancel/focus restoration, and PageState loading/empty/retry/live messages;
- `TransactionForm` type/category filtering, exact POST/PATCH payloads, client request ID lifecycle, invalid amount/date/title focus, and preserved draft after rejected save;
- `ReportBreakdownTable` numeric rows and empty state, and `MoneyText`/`DateText` Indonesian output.

Pinia tests (`tests/stores`) mock only API modules and verify cursor pagination, generation/abort stale-response protection, active-versus-Trash isolation, optimistic version forwarding, mutation invalidation, home zero fallback, custom/preset report requests, auth clear/reset, and preservation of drafts held by the mounted form after request errors.

Playwright tests (`tests/e2e`) run the built Vite UI with a deterministic local API fixture that implements the relevant `/api/v1` browser contract, including `Set-Cookie`, HttpOnly refresh cookie, readable CSRF cookie, token rotation, 401, 409, 422, and delayed responses. Keep backend authorization verification in the backend suite; browser tests assert client request shape and visible behavior.

- `auth.spec.js`: web login request and cookies, refresh after reload, intended-route redirect, failed login retention, CSRF header on refresh/logout, logout, and session-expired redirect.
- `transactions.spec.js`: today/day navigation, exact string create payload, network-failed submit preserves input and retry reuses request ID, edit version, delete confirmation, Trash reload/restore, cursor loading, and no duplicate automatic POST.
- `categories.spec.js`: type-specific create/edit/delete/restore, removed category unavailable to the transaction form, and version conflict copy.
- `reports.spec.js`: 7/30/custom request parameters, invalid inclusive custom range, numeric summary/table/category/period rows, report transaction list, and zero/empty/error states.
- `responsive.spec.js`: keyboard-only critical paths and screenshots/assertions at 360×800 and 1440×900 with no horizontal overflow, visible focus, labels, and readable error/status text.

Run, record, and require these results before declaring the implementation done:

```bash
node --version
npm ci
npx playwright install chromium
npm run test
npm run test:coverage
npm run build
npm run test:e2e
```

For integration evidence, run the API with a disposable PostgreSQL fixture and `AUTH_ALLOWED_ORIGINS` that includes the Vite/Playwright origin, then repeat browser auth and a real personal CRUD/Trash/report path. Separately run the backend commands already configured in `monelog-api`: `make test`, `make vet`, `make test-integration`, and `make check`. Do not claim that UI checks establish backend authorization; backend integration tests must continue to establish cross-user and role denial.

## Risks, compatibility checks, dan exclusions

- The app repository’s three legacy documents describe a different SQLite/USD API. The OpenAPI path, envelopes, parameter names, `isDelete` spelling, `If-Match` quoting, version bodies, cursor behavior, and exact money-string regex are binding. Stop implementation and reconcile any OpenAPI drift before generating client assumptions.
- The current browser auth handler rejects unregistered `Origin` values and requires the CSRF header for refresh/logout. Secure `__Host-` cookies require HTTPS in production; development/proxy and backend origin configuration must be tested early. Do not work around a cookie failure by persisting refresh or access tokens.
- The OpenAPI `ErrorResponse` has code/message only, not field-error maps. Client validation can be field-specific; server 422/409/403/404 must remain a safe form/page-level message with recovery, preserving values. On 409, retain draft and ask the user to reload rather than overwrite another version.
- API totals, dates, ordering, ownership, active/Trash selection, range presets, and currency are server facts. Avoid JavaScript float arithmetic, client total calculation, trusting a route UUID, or rendering data after a changed session/route.
- Screenshots were not stored in the inspected workspaces. Validate final visual spacing and iconography against the supplied source before visual sign-off; their documented functional content is already covered above.

Excluded from this plan: TypeScript; Capacitor/Android/iOS packaging; offline write queue or offline cache; calculator; graph/chart UI; exports; Google Drive; transaction templates; backup/restore; account/role/user administration; admin target selection and all `/api/v1/admin/...` calls; hard delete; and any frontend replacement for backend authorization.
