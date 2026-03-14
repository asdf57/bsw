<?php
$apiBase = getenv('API_BASE') ?: 'http://localhost:8080';
function h(string $v): string { return htmlspecialchars($v, ENT_QUOTES, 'UTF-8'); }
?>
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>BSW</title>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=DM+Sans:ital,opsz,wght@0,9..40,300;0,9..40,400;0,9..40,500;1,9..40,300&family=DM+Mono:wght@400;500&display=swap" rel="stylesheet">
  <script defer src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
  <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.11.1/styles/github.min.css">
  <script src="https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.11.1/highlight.min.js"></script>
  <style>
    *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }

    :root {
      --bg: #f7f6f3;
      --surface: #ffffff;
      --surface-raised: #ffffff;
      --border: #e8e5df;
      --border-strong: #d0ccc4;
      --text: #1a1916;
      --text-2: #6b6760;
      --text-3: #9e9b96;
      --accent: #1a1916;
      --accent-2: #3d3a35;
      --green: #1a6641;
      --green-bg: #edf7f1;
      --red: #991f1f;
      --red-bg: #fdf1f1;
      --amber: #7a4d10;
      --amber-bg: #fef7ed;
      --mono: 'DM Mono', monospace;
      --sans: 'DM Sans', sans-serif;
      --radius: 10px;
      --radius-sm: 6px;
      --shadow: 0 1px 3px rgba(0,0,0,0.06), 0 1px 2px rgba(0,0,0,0.04);
    }

    body {
      font-family: var(--sans);
      background: var(--bg);
      color: var(--text);
      min-height: 100vh;
      font-size: 14px;
      line-height: 1.5;
    }

    .shell {
      max-width: 900px;
      margin: 0 auto;
      padding: 32px 24px 64px;
    }

    .header {
      display: flex;
      align-items: baseline;
      gap: 16px;
      margin-bottom: 40px;
      padding-bottom: 24px;
      border-bottom: 1px solid var(--border);
    }
    .header h1 {
      font-size: 22px;
      font-weight: 500;
      letter-spacing: -0.3px;
    }
    .header .api-badge {
      font-family: var(--mono);
      font-size: 11px;
      color: var(--text-3);
      background: var(--border);
      padding: 3px 8px;
      border-radius: 4px;
    }

    .alert {
      display: flex;
      align-items: center;
      gap: 8px;
      padding: 10px 14px;
      border-radius: var(--radius-sm);
      font-size: 13px;
      margin-bottom: 20px;
    }
    .alert-error { background: var(--red-bg); color: var(--red); border: 1px solid #f5c6c6; }
    .alert-ok    { background: var(--green-bg); color: var(--green); border: 1px solid #b8dfc9; }
    .alert-dot { width: 6px; height: 6px; border-radius: 50%; background: currentColor; flex-shrink: 0; }

    .section { margin-bottom: 32px; }
    .section-header {
      display: flex;
      align-items: center;
      justify-content: space-between;
      margin-bottom: 14px;
    }
    .section-header h2 {
      font-size: 13px;
      font-weight: 500;
      text-transform: uppercase;
      letter-spacing: 0.06em;
      color: var(--text-3);
    }
    .section-actions {
      display: flex;
      align-items: center;
      gap: 8px;
    }

    .card {
      background: var(--surface);
      border: 1px solid var(--border);
      border-radius: var(--radius);
      overflow: hidden;
      box-shadow: var(--shadow);
    }

    .add-form {
      display: flex;
      gap: 8px;
      flex-wrap: wrap;
      align-items: center;
      padding: 14px 16px;
      background: var(--bg);
      border-bottom: 1px solid var(--border);
    }
    .add-form input {
      font-family: var(--sans);
      font-size: 13px;
      padding: 7px 11px;
      border: 1px solid var(--border-strong);
      border-radius: var(--radius-sm);
      background: var(--surface);
      color: var(--text);
      outline: none;
      transition: border-color 0.15s;
      min-width: 100px;
    }
    .add-form input:focus { border-color: var(--accent); }
    .add-form input::placeholder { color: var(--text-3); }
    .add-form input[type="number"] { width: 100px; }
    .add-form .input-wide { flex: 1; min-width: 160px; }

    .btn {
      font-family: var(--sans);
      font-size: 13px;
      font-weight: 500;
      padding: 7px 14px;
      border-radius: var(--radius-sm);
      cursor: pointer;
      border: 1px solid transparent;
      transition: all 0.15s;
      white-space: nowrap;
      line-height: 1;
    }
    .btn-primary {
      background: var(--accent);
      color: #fff;
      border-color: var(--accent);
    }
    .btn-primary:hover { background: var(--accent-2); border-color: var(--accent-2); }
    .btn-ghost {
      background: transparent;
      color: var(--text-2);
      border-color: var(--border-strong);
    }
    .btn-ghost:hover { background: var(--bg); color: var(--text); }
    .btn-danger {
      background: transparent;
      color: var(--red);
      border-color: transparent;
      padding: 4px 8px;
      font-size: 12px;
    }
    .btn-danger:hover { background: var(--red-bg); border-color: #f5c6c6; }

    .data-table { width: 100%; border-collapse: collapse; }
    .table-scroll {
      overflow-x: auto;
      -webkit-overflow-scrolling: touch;
    }
    .table-scroll .data-table { min-width: 700px; }
    .data-table th {
      font-size: 11px;
      font-weight: 500;
      text-transform: uppercase;
      letter-spacing: 0.06em;
      color: var(--text-3);
      padding: 10px 16px;
      text-align: left;
      background: var(--bg);
      border-bottom: 1px solid var(--border);
    }
    .data-table td {
      padding: 11px 16px;
      border-bottom: 1px solid var(--border);
      vertical-align: middle;
      font-size: 13px;
      color: var(--text);
    }
    .data-table tbody tr:last-child td { border-bottom: none; }
    .data-table tbody tr:hover td { background: #fafaf8; }

    .cell-id { font-family: var(--mono); font-size: 11px; color: var(--text-3); }
    .cell-date { font-size: 12px; color: var(--text-3); }
    .cell-amount { font-family: var(--mono); font-weight: 500; font-size: 13px; }
    .empty-row td { text-align: center; color: var(--text-3); padding: 32px 16px; font-size: 13px; }

    .balance-from { font-weight: 500; }
    .balance-arrow { color: var(--text-3); margin: 0 6px; }
    .balance-to { color: var(--text-2); }
    .balance-amount {
      font-family: var(--mono);
      font-weight: 500;
      color: var(--amber);
      background: var(--amber-bg);
      padding: 3px 8px;
      border-radius: 4px;
      font-size: 12px;
    }

    .view-toggle {
      position: relative;
      display: inline-flex;
      align-items: center;
      padding: 2px;
      border: 1px solid var(--border);
      border-radius: 999px;
      background: var(--bg);
      gap: 2px;
    }
    .view-toggle-slider {
      position: absolute;
      top: 2px;
      bottom: 2px;
      width: calc(50% - 2px);
      border-radius: 999px;
      background: var(--surface);
      box-shadow: 0 1px 2px rgba(15, 23, 42, 0.08);
      transition: transform 0.18s ease;
    }
    .view-toggle-slider.is-code { transform: translateX(100%); }
    .view-toggle-btn {
      position: relative;
      z-index: 1;
      min-width: 72px;
      padding: 6px 12px;
      border: none;
      background: transparent;
      border-radius: 999px;
      font-family: var(--sans);
      font-size: 12px;
      color: var(--text-3);
      cursor: pointer;
      transition: color 0.15s ease;
    }
    .view-toggle-btn.is-active {
      color: var(--text);
      font-weight: 500;
    }
    .code-block {
      margin: 0;
      padding: 16px;
      overflow-x: auto;
      border-radius: var(--radius-sm);
      background: #fafaf8;
      border: 1px solid var(--border);
      color: var(--text);
      font-family: var(--mono);
      font-size: 12px;
      line-height: 1.55;
      white-space: pre-wrap;
      word-break: break-word;
    }
    .code-panel {
      position: relative;
    }
    .code-block code {
      display: block;
      font-family: inherit;
      font-size: inherit;
      line-height: inherit;
      color: inherit;
      background: transparent;
    }
    .code-download-btn {
      position: absolute;
      top: 10px;
      right: 10px;
      z-index: 2;
      font-family: var(--sans);
      font-size: 11px;
      color: var(--text-3);
      background: rgba(255, 255, 255, 0.92);
      border: 1px solid var(--border);
      border-radius: 999px;
      padding: 5px 10px;
      cursor: pointer;
      transition: all 0.15s ease;
      backdrop-filter: blur(2px);
    }
    .code-download-btn:hover {
      color: var(--text);
      border-color: var(--border-strong);
      background: #fcfcfa;
    }

    .refresh-btn {
      font-family: var(--sans);
      font-size: 12px;
      color: var(--text-3);
      background: none;
      border: none;
      cursor: pointer;
      padding: 4px 8px;
      border-radius: var(--radius-sm);
      transition: all 0.15s;
    }
    .refresh-btn:hover { background: var(--border); color: var(--text); }
    .header-download-btn {
      font-family: var(--sans);
      font-size: 12px;
      color: var(--text-3);
      background: none;
      border: none;
      cursor: pointer;
      padding: 4px 8px;
      border-radius: var(--radius-sm);
      transition: all 0.15s;
    }
    .header-download-btn:hover { background: var(--border); color: var(--text); }
  </style>
</head>
<body x-data="app()" x-init="init()">

  <!-- Modal -->
  <div x-show="modal.open" @keydown.escape.window="modal.open = false" style="position: fixed; inset: 0; background: rgba(0,0,0,0.4); z-index: 50;">
    <div @click.self="modal.open = false" style="display: flex; align-items: center; justify-content: center; width: 100%; height: 100%;">
      <div style="background: var(--surface); border: 1px solid var(--border); border-radius: var(--radius); box-shadow: 0 8px 32px rgba(0,0,0,0.12); padding: 24px; width: 100%; max-width: 360px; margin: 0 16px;">
        <p style="font-size: 14px; color: var(--text); margin-bottom: 20px;" x-text="modal.message"></p>
        <div style="display: flex; gap: 8px; justify-content: flex-end;">
          <button class="btn btn-ghost" @click="modal.open = false">Cancel</button>
          <button class="btn btn-danger" style="border-color: #f5c6c6; padding: 7px 14px; font-size: 13px;" @click="modal.onConfirm(); modal.open = false">Delete</button>
        </div>
      </div>
    </div>
  </div>

  <div class="shell">

    <header class="header">
      <h1>BSW</h1>
      <span class="api-badge"><?= h(rtrim($apiBase, '/')) ?></span>
    </header>

    <div class="alert alert-error" x-show="error" x-transition>
      <span class="alert-dot"></span>
      <span x-text="error"></span>
    </div>
    <div class="alert alert-ok" x-show="notice" x-transition>
      <span class="alert-dot"></span>
      <span x-text="notice"></span>
    </div>

    <!-- Users -->
    <div class="section">
      <div class="section-header">
        <h2>Users</h2>
        <button class="refresh-btn" @click="loadAll()">↻ Refresh</button>
      </div>
      <div class="card">
        <form class="add-form" @submit.prevent="createUser()">
          <input x-model.trim="newUserName" placeholder="Name" required>
          <button type="submit" class="btn btn-primary">Add user</button>
        </form>
        <table class="data-table">
          <thead>
            <tr><th>ID</th><th>Name</th><th>Created</th><th></th></tr>
          </thead>
          <tbody>
            <template x-if="users.length === 0">
              <tr class="empty-row"><td colspan="4">No users yet</td></tr>
            </template>
            <template x-for="u in users" :key="u.ID">
              <tr>
                <td class="cell-id" x-text="u.ID"></td>
                <td x-text="u.Name"></td>
                <td class="cell-date" x-text="u.CreatedAt"></td>
                <td><button class="btn btn-danger" @click="deleteUser(u.Name)">Remove</button></td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Payments -->
    <div class="section">
      <div class="section-header">
        <h2>Payments</h2>
        <div class="section-actions">
          <button type="button" class="header-download-btn" @click="downloadPayments()">Download</button>
        </div>
      </div>
      <div class="card">
        <form class="add-form" @submit.prevent="createPayment()">
          <input x-model.number="newPayment.amount" type="number" step="0.01" min="0" placeholder="Amount" required>
          <input x-model.trim="newPayment.payer" placeholder="Payer" required>
          <input x-model.trim="newPayment.description" class="input-wide" placeholder="Description" required>
          <input x-model.trim="newPayment.owersRaw" class="input-wide" placeholder="Owers (comma-separated)">
          <button type="submit" class="btn btn-primary">Add payment</button>
        </form>
        <div class="table-scroll">
          <table class="data-table">
            <thead>
              <tr><th>ID</th><th>Payer</th><th>Owers</th><th>Amount</th><th>Description</th><th>Date</th><th></th></tr>
            </thead>
            <tbody>
              <template x-if="payments.length === 0">
                <tr class="empty-row"><td colspan="6">No payments yet</td></tr>
              </template>
              <template x-for="p in payments" :key="p.ID">
                <tr>
                  <td class="cell-id" x-text="p.ID"></td>
                  <td class="cell-id" x-text="p.PayerName"></td>
                  <td class="cell-id" x-text="p.Owers"></td>
                  <td class="cell-amount" x-text="p.Amount"></td>
                  <td x-text="p.Description"></td>
                  <td class="cell-date" x-text="p.Date"></td>
                  <td><button class="btn btn-danger" @click="deletePayment(p.ID)">Remove</button></td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <!-- Balances -->
    <div class="section">
      <div class="section-header">
        <h2>Balances</h2>
        <div class="view-toggle" role="tablist" aria-label="Balances view mode">
          <div
            class="view-toggle-slider"
            :class="{ 'is-code': balancesView === 'code' }"
            aria-hidden="true"
          ></div>
          <button
            type="button"
            class="view-toggle-btn"
            :class="{ 'is-active': balancesView === 'table' }"
            :aria-selected="balancesView === 'table'"
            @click="balancesView = 'table'"
          >Table</button>
          <button
            type="button"
            class="view-toggle-btn"
            :class="{ 'is-active': balancesView === 'code' }"
            :aria-selected="balancesView === 'code'"
            @click="showBalancesCode()"
          >Code</button>
        </div>
      </div>
      <div class="card">
        <template x-if="balancesView === 'table'">
          <table class="data-table">
            <thead>
              <tr><th>Who owes whom</th><th>Amount</th></tr>
            </thead>
            <tbody>
              <template x-if="balances.length === 0">
                <tr class="empty-row"><td colspan="2">All settled up</td></tr>
              </template>
              <template x-for="b in balances" :key="`${b.FromUser}-${b.ToUser}`">
                <tr>
                  <td>
                    <span class="balance-from" x-text="b.FromUser"></span>
                    <span class="balance-arrow">→</span>
                    <span class="balance-to" x-text="b.ToUser"></span>
                  </td>
                  <td><span class="balance-amount" x-text="b.Amount"></span></td>
                </tr>
              </template>
            </tbody>
          </table>
        </template>
        <template x-if="balancesView === 'code'">
          <div class="code-panel">
            <button type="button" class="code-download-btn" @click="downloadBalances()">Download</button>
            <pre class="code-block"><code x-ref="balancesCode" class="language-json"></code></pre>
          </div>
        </template>
      </div>
    </div>

  </div><!-- /shell -->

  <script>
    function app() {
      const apiBase = "<?= h(rtrim($apiBase, '/')) ?>";
      const parseError = async (resp) => {
        try {
          const j = await resp.json();
          return j.error || JSON.stringify(j);
        } catch (_) {
          return `HTTP ${resp.status}`;
        }
      };
      const req = async (path, opts = {}) => {
        const resp = await fetch(apiBase + path, {
          headers: { "Content-Type": "application/json" },
          ...opts,
        });
        if (!resp.ok) throw new Error(await parseError(resp));
        if (resp.status === 204) return null;
        const t = await resp.text();
        return t ? JSON.parse(t) : null;
      };
      const reqWithRaw = async (path, opts = {}) => {
        const resp = await fetch(apiBase + path, {
          headers: { "Content-Type": "application/json" },
          ...opts,
        });
        if (!resp.ok) throw new Error(await parseError(resp));
        if (resp.status === 204) return { data: null, raw: "null" };
        const raw = await resp.text();
        const data = raw ? JSON.parse(raw) : null;
        return {
          data,
          raw: JSON.stringify(data, null, 2),
        };
      };

      return {
        users: [],
        payments: [],
        balances: [],
        balancesRaw: "[]",
        balancesView: "table",
        newUserName: "",
        newPayment: { amount: "", payer: "", description: "", owersRaw: "" },
        modal: { open: false, message: "", onConfirm: null },
        error: "",
        notice: "",

        async init() {
          await this.loadAll();
          setInterval(() => this.loadAll(true), 10000);
        },

        flashError(msg) {
          this.error = msg;
          if (msg) this.notice = "";
        },

        flashNotice(msg) {
          this.notice = msg;
          if (msg) this.error = "";
          setTimeout(() => { if (this.notice === msg) this.notice = ""; }, 3000);
        },

        showBalancesCode() {
          this.balancesView = "code";
          this.highlightBalances();
        },

        highlightBalances() {
          this.$nextTick(() => {
            if (!this.$refs.balancesCode || typeof hljs === "undefined") return;
            delete this.$refs.balancesCode.dataset.highlighted;
            this.$refs.balancesCode.classList.remove("hljs");
            this.$refs.balancesCode.textContent = this.balancesRaw;
            hljs.highlightElement(this.$refs.balancesCode);
          });
        },

        downloadBalances() {
          const blob = new Blob([this.balancesRaw], { type: "application/json" });
          const url = URL.createObjectURL(blob);
          const a = document.createElement("a");
          const stamp = new Date().toISOString().replace(/[:.]/g, "-");
          a.href = url;
          a.download = `balances-${stamp}.json`;
          document.body.appendChild(a);
          a.click();
          a.remove();
          URL.revokeObjectURL(url);
        },

        downloadPayments() {
          const data = JSON.stringify(this.payments, null, 2);
          const blob = new Blob([data], { type: "application/json" });
          const url = URL.createObjectURL(blob);
          const a = document.createElement("a");
          const stamp = new Date().toISOString().replace(/[:.]/g, "-");
          a.href = url;
          a.download = `payments-${stamp}.json`;
          document.body.appendChild(a);
          a.click();
          a.remove();
          URL.revokeObjectURL(url);
        },

        confirm(message, fn) {
          this.modal = { open: true, message, onConfirm: fn };
        },

        async loadAll(background = false) {
          try {
            // update balances
            await req("/api/v1/balance/all", {
              method: "POST",
            });

            const [users, payments, balancesResp] = await Promise.all([
              req("/api/v1/user/all"),
              req("/api/v1/payment/all"),
              reqWithRaw("/api/v1/balance/all")
            ]);
            this.users    = Array.isArray(users)    ? users    : [];
            this.payments = Array.isArray(payments) ? payments : [];
            this.balances = Array.isArray(balancesResp.data) ? balancesResp.data : [];
            this.balancesRaw = balancesResp.raw;
            if (this.balancesView === "code") this.highlightBalances();
            if (!background) this.flashError("");
          } catch (e) {
            this.flashError(`Load failed: ${e.message}`);
          }
        },

        async createUser() {
          this.flashError("");
          try {
            await req("/api/v1/user", {
              method: "POST",
              body: JSON.stringify({ name: this.newUserName }),
            });
            this.newUserName = "";
            this.flashNotice("User created");
            await this.loadAll();
          } catch (e) {
            this.flashError(`Create user failed: ${e.message}`);
          }
        },

        async deleteUser(name) {
          if (!name) { this.flashError("Invalid user name"); return; }
          this.confirm(`Delete user "${name}"?`, async () => {
            this.flashError("");
            try {
              await req(`/api/v1/user/${encodeURIComponent(name)}`, { method: "DELETE" });
              this.flashNotice("User deleted");
              await this.loadAll();
            } catch (e) {
              this.flashError(`Delete user failed: ${e.message}`);
            }
          });
        },

        async createPayment() {
          this.flashError("");
          const owers = this.newPayment.owersRaw
            .split(",")
            .map((s) => s.trim())
            .filter(Boolean);
          try {
            await req("/api/v1/payment", {
              method: "POST",
              body: JSON.stringify({
                amount:      Number(this.newPayment.amount),
                payer:       this.newPayment.payer,
                description: this.newPayment.description,
                owers,
              }),
            });
            this.newPayment = { amount: "", payer: "", description: "", owersRaw: "" };
            this.flashNotice("Payment created");
            await this.loadAll();
          } catch (e) {
            this.flashError(`Create payment failed: ${e.message}`);
          }
        },

        async deletePayment(id) {
          if (!id) { this.flashError("Invalid payment ID"); return; }
          this.confirm(`Delete payment #${id}?`, async () => {
            this.flashError("");
            try {
              await req(`/api/v1/payment/${id}`, { method: "DELETE" });
              this.flashNotice("Payment deleted");
              await this.loadAll();
            } catch (e) {
              this.flashError(`Delete payment failed: ${e.message}`);
            }
          });
        },
      };
    }
  </script>
</body>
</html>
