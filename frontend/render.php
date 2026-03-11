<?php
$apiBase = getenv('API_BASE') ?: 'http://localhost:8080';
function h(string $v): string { return htmlspecialchars($v, ENT_QUOTES, 'UTF-8'); }
?>
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Better Splitwise</title>
  <script defer src="https://unpkg.com/alpinejs@3.x.x/dist/cdn.min.js"></script>
  <style>
    body { font-family: Arial, sans-serif; margin: 24px; color: #111; }
    section { margin-top: 20px; }
    h1 { margin: 0 0 6px; }
    form { display: flex; gap: 8px; flex-wrap: wrap; margin: 8px 0 12px; }
    input { padding: 6px 8px; border: 1px solid #ccc; border-radius: 4px; }
    button { padding: 6px 10px; border: 1px solid #bbb; background: #f5f5f5; border-radius: 4px; cursor: pointer; }
    button.danger { border-color: #c11; color: #c11; background: #fff; }
    table { border-collapse: collapse; width: 100%; max-width: 1000px; }
    th, td { border: 1px solid #ddd; padding: 8px; text-align: left; vertical-align: top; }
    th { background: #f7f7f7; }
    .muted { color: #666; }
    .error { color: #b00020; }
    .ok { color: #126a2f; }
  </style>
</head>
<body x-data="app()" x-init="init()">
  <h1>BSW</h1>
  <div class="muted">API: <?= h(rtrim($apiBase, '/')) ?></div>
  <div class="ok" x-text="notice" x-show="notice"></div>
  <div class="error" x-text="error" x-show="error"></div>

  <section>
    <h2>Users</h2>
    <form @submit.prevent="createUser()">
      <input x-model.trim="newUserName" placeholder="Name" required>
      <button type="submit">Add User</button>
      <button type="button" @click="loadAll()">Refresh</button>
    </form>
    <table>
      <thead><tr><th>ID</th><th>Name</th><th>Created</th><th>Action</th></tr></thead>
      <tbody>
        <template x-if="users.length === 0"><tr><td colspan="4" class="muted">No users found.</td></tr></template>
        <template x-for="u in users" :key="u.ID">
          <tr>
            <td x-text="u.ID"></td>
            <td x-text="u.Name"></td>
            <td x-text="u.CreatedAt"></td>
            <td><button class="danger" @click="deleteUser(u.Name)">Delete</button></td>
          </tr>
        </template>
      </tbody>
    </table>
  </section>

  <section>
    <h2>Payments</h2>
    <form @submit.prevent="createPayment()">
      <input x-model.number="newPayment.amount" type="number" step="0.01" min="0" placeholder="Amount" required>
      <input x-model.trim="newPayment.payer" placeholder="Payer name" required>
      <input x-model.trim="newPayment.description" placeholder="Description" required>
      <input x-model.trim="newPayment.owersRaw" placeholder="Owers (comma separated)">
      <button type="submit">Add Payment</button>
    </form>
    <table>
      <thead><tr><th>ID</th><th>Amount</th><th>Description</th><th>Payer ID</th><th>Date</th><th>Action</th></tr></thead>
      <tbody>
        <template x-if="payments.length === 0"><tr><td colspan="6" class="muted">No payments found.</td></tr></template>
        <template x-for="p in payments" :key="p.ID">
          <tr>
            <td x-text="p.ID"></td>
            <td x-text="p.Amount"></td>
            <td x-text="p.Description"></td>
            <td x-text="p.PayerID"></td>
            <td x-text="p.Date"></td>
            <td><button class="danger" @click="deletePayment(p.ID)">Delete</button></td>
          </tr>
        </template>
      </tbody>
    </table>
  </section>

  <section>
    <h2>Balances</h2>
      <table>
        <thead><tr><th>Debt</th><th>Amount</th></tr></thead>
        <tbody>
          <template x-if="balances.length === 0"><tr><td colspan="2" class="muted">No debts found.</td></tr></template>
          <template x-for="b in balances" :key="`${b.FromUser}-${b.ToUser}`">
            <tr>
              <td x-text="`${b.FromUser} owes ${b.ToUser}`"></td>
              <td x-text="b.Amount"></td>
            </tr>
          </template>
        </tbody>
      </table>

  </section>

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

      return {
        users: [],
        payments: [],
        balances: [],
        newUserName: "",
        newPayment: { amount: "", payer: "", description: "", owersRaw: "" },
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
        },

        async loadAll(background = false) {
          try {
            const [users, payments, balances] = await Promise.all([
              req("/api/v1/user/all"),
              req("/api/v1/payment/all"),
              req("/api/v1/balance/all")
            ]);
            this.users = Array.isArray(users) ? users : [];
            this.payments = Array.isArray(payments) ? payments : [];
            this.balances = Array.isArray(balances) ? balances : [];
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
          if (!name || !confirm(`Delete user ${name}?`)) return;
          this.flashError("");
          try {
            await req(`/api/v1/user/${encodeURIComponent(name)}`, { method: "DELETE" });
            this.flashNotice("User deleted");
            await this.loadAll();
          } catch (e) {
            this.flashError(`Delete user failed: ${e.message}`);
          }
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
                amount: Number(this.newPayment.amount),
                payer: this.newPayment.payer,
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
          if (!id || !confirm(`Delete payment ${id}?`)) return;
          this.flashError("");
          try {
            await req(`/api/v1/payment/${id}`, { method: "DELETE" });
            this.flashNotice("Payment deleted");
            await this.loadAll();
          } catch (e) {
            this.flashError(`Delete payment failed: ${e.message}`);
          }
        },
      };
    }
  </script>
</body>
</html>
