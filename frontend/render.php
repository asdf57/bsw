<?php
$apiBase = getenv("API_BASE") ?: "http://localhost:8080";
function h(string $v): string
{
    return htmlspecialchars($v, ENT_QUOTES, "UTF-8");
}
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
    .card-payments {
      overflow: visible;
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
    .add-form input:not([type="checkbox"]) {
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
    .add-form select {
      font-family: var(--sans);
      font-size: 13px;
      padding: 7px 34px 7px 11px;
      border: 1px solid var(--border-strong);
      border-radius: var(--radius-sm);
      background-color: var(--surface);
      background-image:
        linear-gradient(45deg, transparent 50%, var(--text-3) 50%),
        linear-gradient(135deg, var(--text-3) 50%, transparent 50%);
      background-position:
        calc(100% - 16px) calc(50% - 1px),
        calc(100% - 11px) calc(50% - 1px);
      background-size: 5px 5px, 5px 5px;
      background-repeat: no-repeat;
      color: var(--text);
      outline: none;
      transition: border-color 0.15s;
      min-width: 140px;
      appearance: none;
      -webkit-appearance: none;
      -moz-appearance: none;
    }
    .add-form input:not([type="checkbox"]):focus { border-color: var(--accent); }
    .add-form select:focus { border-color: var(--accent); }
    .add-form input:not([type="checkbox"])::placeholder { color: var(--text-3); }
    .add-form select:invalid { color: var(--text-3); }
    .add-form select option { color: var(--text); }
    .add-form input[type="number"] { width: 100px; }
    .add-form .input-wide { flex: 1; min-width: 160px; }
    .owers-dropdown {
      position: relative;
      min-width: 220px;
    }
    .payer-dropdown {
      position: relative;
      min-width: 220px;
    }
    .payer-trigger {
      width: 100%;
      height: 33px;
      padding: 7px 34px 7px 11px;
      border: 1px solid var(--border-strong);
      border-radius: var(--radius-sm);
      background-color: var(--surface);
      color: var(--text);
      font-family: var(--sans);
      font-size: 13px;
      text-align: left;
      cursor: pointer;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
      transition: border-color 0.15s;
      background-image:
        linear-gradient(45deg, transparent 50%, var(--text-3) 50%),
        linear-gradient(135deg, var(--text-3) 50%, transparent 50%);
      background-position:
        calc(100% - 16px) calc(50% - 1px),
        calc(100% - 11px) calc(50% - 1px);
      background-size: 5px 5px, 5px 5px;
      background-repeat: no-repeat;
    }
    .payer-trigger.is-open { border-color: var(--accent); }
    .payer-trigger.is-placeholder { color: var(--text-3); }
    .payer-menu {
      position: absolute;
      top: calc(100% + 6px);
      left: 0;
      right: 0;
      z-index: 60;
      max-height: 180px;
      overflow-y: auto;
      -webkit-overflow-scrolling: touch;
      padding: 6px 0;
      border: 1px solid var(--border-strong);
      border-radius: var(--radius-sm);
      background: var(--surface);
      box-shadow: var(--shadow);
    }
    .payer-option {
      width: 100%;
      border: none;
      background: transparent;
      color: var(--text);
      font-family: var(--sans);
      font-size: 13px;
      text-align: left;
      padding: 6px 10px;
      cursor: pointer;
    }
    .payer-option:hover { background: var(--bg); }
    .payer-option.is-active {
      background: var(--bg);
      font-weight: 500;
    }
    .owers-trigger {
      width: 100%;
      height: 33px;
      padding: 7px 34px 7px 11px;
      border: 1px solid var(--border-strong);
      border-radius: var(--radius-sm);
      background-color: var(--surface);
      color: var(--text);
      font-family: var(--sans);
      font-size: 13px;
      text-align: left;
      cursor: pointer;
      white-space: nowrap;
      overflow: hidden;
      text-overflow: ellipsis;
      transition: border-color 0.15s;
      background-image:
        linear-gradient(45deg, transparent 50%, var(--text-3) 50%),
        linear-gradient(135deg, var(--text-3) 50%, transparent 50%);
      background-position:
        calc(100% - 16px) calc(50% - 1px),
        calc(100% - 11px) calc(50% - 1px);
      background-size: 5px 5px, 5px 5px;
      background-repeat: no-repeat;
    }
    .owers-trigger.is-open { border-color: var(--accent); }
    .owers-menu {
      position: absolute;
      top: calc(100% + 6px);
      left: 0;
      right: 0;
      z-index: 60;
      max-height: 180px;
      overflow-y: auto;
      -webkit-overflow-scrolling: touch;
      padding: 6px 0;
      border: 1px solid var(--border-strong);
      border-radius: var(--radius-sm);
      background: var(--surface);
      box-shadow: var(--shadow);
    }
    .owers-option {
      display: grid;
      grid-template-columns: 13px 1fr;
      align-items: center;
      justify-content: start;
      column-gap: 10px;
      width: 100%;
      font-size: 13px;
      color: var(--text);
      padding: 6px 10px;
      cursor: pointer;
      user-select: none;
    }
    .owers-option:hover {
      background: var(--bg);
    }
    .owers-option:has(input[type="checkbox"]:checked) {
      background: var(--bg);
    }
    .owers-option input[type="checkbox"] {
      width: 13px;
      height: 13px;
      margin: 0;
      appearance: none;
      -webkit-appearance: none;
      border: 1px solid var(--border-strong);
      border-radius: 3px;
      background: var(--surface);
      display: inline-grid;
      place-content: center;
      cursor: pointer;
    }
    .owers-option input[type="checkbox"]::before {
      content: "";
      width: 7px;
      height: 7px;
      transform: scale(0);
      transition: transform 0.12s ease-in-out;
      box-shadow: inset 1em 1em #fff;
      clip-path: polygon(14% 44%, 0 60%, 45% 100%, 100% 18%, 82% 0, 43% 62%);
    }
    .owers-option input[type="checkbox"]:checked {
      background: var(--accent);
      border-color: var(--accent);
    }
    .owers-option input[type="checkbox"]:checked::before {
      transform: scale(1);
    }
    .owers-option input[type="checkbox"]:focus-visible {
      outline: 2px solid var(--border-strong);
      outline-offset: 1px;
    }
    .owers-option span {
      text-align: left;
    }
    .owers-empty {
      font-size: 12px;
      color: var(--text-3);
    }

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
    .data-table th.sortable {
      cursor: pointer;
      position: relative;
      user-select: none;
      padding-right: 18px;
    }
    .sort-indicator {
      position: absolute;
      right: 16px;
      top: 50%;
      transform: translateY(-50%);
      font-size: 10px;
      color: var(--text-3);
      text-transform: none;
    }
    .col-resize-handle {
      position: absolute;
      top: 0;
      right: -4px;
      width: 8px;
      height: 100%;
      cursor: col-resize;
      z-index: 2;
    }
    .col-resize-handle::after {
      content: "";
      position: absolute;
      right: 3px;
      top: 20%;
      bottom: 20%;
      width: 1px;
      background: var(--border);
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
    .amount-header { text-align: center !important; }
    .payment-amount-cell { text-align: center; white-space: nowrap; }
    .payment-amount-group {
      display: inline-flex;
      align-items: center;
      justify-content: center;
    }
    .amount-btn {
      background: none;
      border: none;
      padding: 0;
      margin: 0;
      cursor: pointer;
      color: inherit;
      font: inherit;
    }
    .amount-btn:hover .cell-amount {
      text-decoration: underline;
      text-underline-offset: 2px;
    }
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
    .table-pager {
      display: grid;
      grid-template-columns: auto 300px auto;
      align-items: center;
      justify-content: center;
      gap: 12px;
      padding: 10px 16px;
      border-top: 1px solid var(--border);
    }
    .table-pager-info {
      font-size: 12px;
      color: var(--text-3);
      text-align: center;
      white-space: nowrap;
      font-variant-numeric: tabular-nums;
    }
    .table-pager .btn:disabled {
      opacity: 0.5;
      cursor: not-allowed;
    }
    @media (max-width: 640px) {
      .table-pager {
        grid-template-columns: 1fr;
      }
      .table-pager-info {
        white-space: normal;
      }
    }
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
      <span class="api-badge"><?= h(rtrim($apiBase, "/")) ?></span>
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
          <colgroup>
            <col :style="`width:${tableUi.users.widths.id}px`">
            <col :style="`width:${tableUi.users.widths.name}px`">
            <col :style="`width:${tableUi.users.widths.created}px`">
            <col :style="`width:${tableUi.users.widths.actions}px`">
          </colgroup>
          <thead>
            <tr>
              <th class="sortable" @click="toggleTableSort('users', 'ID')">ID <span class="sort-indicator" x-text="tableSortIndicator('users', 'ID')"></span><span class="col-resize-handle" @mousedown.stop.prevent="startColumnResize('users', 'id', $event)"></span></th>
              <th class="sortable" @click="toggleTableSort('users', 'Name')">Name <span class="sort-indicator" x-text="tableSortIndicator('users', 'Name')"></span><span class="col-resize-handle" @mousedown.stop.prevent="startColumnResize('users', 'name', $event)"></span></th>
              <th class="sortable" @click="toggleTableSort('users', 'CreatedAt')">Created <span class="sort-indicator" x-text="tableSortIndicator('users', 'CreatedAt')"></span><span class="col-resize-handle" @mousedown.stop.prevent="startColumnResize('users', 'created', $event)"></span></th>
              <th><span class="col-resize-handle" @mousedown.stop.prevent="startColumnResize('users', 'actions', $event)"></span></th>
            </tr>
          </thead>
          <tbody>
            <template x-if="users.length === 0">
              <tr class="empty-row"><td colspan="4">No users yet</td></tr>
            </template>
            <template x-for="u in getTablePageRows('users', users)" :key="u.ID">
              <tr>
                <td class="cell-id" x-text="u.ID"></td>
                <td x-text="u.Name"></td>
                <td class="cell-date" x-text="u.CreatedAt"></td>
                <td><button class="btn btn-danger" @click="deleteUser(u.Name)">Remove</button></td>
              </tr>
            </template>
          </tbody>
        </table>
        <div class="table-pager" x-show="getTablePageCount('users', users.length) > 1">
          <button
            type="button"
            class="btn btn-ghost"
            @click="viewTablePage('users', tableUi.users.page - 1, users.length)"
            :disabled="tableUi.users.page === 0"
          >Prev</button>
          <div class="table-pager-info">
            Page <span x-text="tableUi.users.page + 1"></span> of
            <span x-text="getTablePageCount('users', users.length)"></span> | Showing
            <span x-text="getTablePageStart('users', users.length)"></span> to
            <span x-text="getTablePageEnd('users', users.length)"></span>
          </div>
          <button
            type="button"
            class="btn btn-ghost"
            @click="viewTablePage('users', tableUi.users.page + 1, users.length)"
            :disabled="tableUi.users.page >= getTablePageCount('users', users.length) - 1"
          >Next</button>
        </div>
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
      <div class="card card-payments">
        <form class="add-form" @submit.prevent="createPayment()">
          <input x-model.number="newPayment.amount" type="number" step="0.01" min="0" placeholder="Amount" required>
          <div class="payer-dropdown" @click.outside="payerOpen = false">
            <button
              type="button"
              class="payer-trigger"
              :class="{ 'is-open': payerOpen, 'is-placeholder': !newPayment.payer }"
              @click="payerOpen = !payerOpen"
              x-text="newPayment.payer || 'Select Payer'"
            ></button>
            <div class="payer-menu" x-show="payerOpen" x-transition>
              <template x-if="users.length === 0">
                <span class="owers-empty">Add users to select payer</span>
              </template>
              <template x-for="u in users" :key="`payer-${u.ID}`">
                <button
                  type="button"
                  class="payer-option"
                  :class="{ 'is-active': newPayment.payer === u.Name }"
                  @click="newPayment.payer = u.Name; newPayment.owers = newPayment.owers.filter((name) => name !== u.Name); payerOpen = false"
                  x-text="u.Name"
                ></button>
              </template>
            </div>
          </div>
          <input x-model.trim="newPayment.description" class="input-wide" placeholder="Description" required>

          <div class="owers-dropdown" @click.outside="owersOpen = false">
            <button
              type="button"
              class="owers-trigger"
              :class="{ 'is-open': owersOpen }"
              @click="owersOpen = !owersOpen"
              x-text="newPayment.owers.length ? `${newPayment.owers.length} ower${newPayment.owers.length > 1 ? 's' : ''} selected` : 'Select Owers'"
            ></button>
            <div class="owers-menu" x-show="owersOpen" x-transition>
              <template x-if="users.filter((u) => u.Name !== newPayment.payer).length === 0">
                <span class="owers-empty">Add users to select owers</span>
              </template>
              <template x-for="u in users.filter((u) => u.Name !== newPayment.payer)" :key="`ower-${u.ID}`">
                <label class="owers-option">
                  <input type="checkbox" x-model="newPayment.owers" :value="u.Name">
                  <span x-text="u.Name"></span>
                </label>
              </template>
            </div>
          </div>

          <div class="payer-dropdown" @click.outside="fromExchangeRateOpen = false">
            <button
              type="button"
              class="payer-trigger"
              :class="{ 'is-open': fromExchangeRateOpen, 'is-placeholder': !newPayment.fromExchangeRate }"
              @click="fromExchangeRateOpen = !fromExchangeRateOpen"
              x-text="newPayment.fromExchangeRate || 'From Currency'"
            ></button>
            <div class="payer-menu" x-show="fromExchangeRateOpen" x-transition>
              <template x-for="e in exchangeRates" :key="`exchange-rate-${e}`">
                <button
                  type="button"
                  class="payer-option"
                  :class="{ 'is-active': newPayment.fromExchangeRate === e }"
                  @click="newPayment.fromExchangeRate = e; fromExchangeRateOpen = false"
                  x-text="e"
                ></button>
              </template>
            </div>
          </div>

          <div class="payer-dropdown" @click.outside="toExchangeRateOpen = false">
            <button
              type="button"
              class="payer-trigger"
              :class="{ 'is-open': toExchangeRateOpen, 'is-placeholder': !newPayment.toExchangeRate }"
              @click="toExchangeRateOpen = !toExchangeRateOpen"
              x-text="newPayment.toExchangeRate || 'To Currency'"
            ></button>
            <div class="payer-menu" x-show="toExchangeRateOpen" x-transition>
              <template x-for="e in exchangeRates" :key="`exchange-rate-${e}`">
                <button
                  type="button"
                  class="payer-option"
                  :class="{ 'is-active': newPayment.toExchangeRate === e }"
                  @click="newPayment.toExchangeRate = e; toExchangeRateOpen = false"
                  x-text="e"
                ></button>
              </template>
            </div>
          </div>

          <button type="submit" class="btn btn-primary">Add payment</button>
        </form>

        <div class="table-scroll">
          <table class="data-table">
            <colgroup>
              <col :style="`width:${tableUi.payments.widths.id}px`">
              <col :style="`width:${tableUi.payments.widths.payer}px`">
              <col :style="`width:${tableUi.payments.widths.owers}px`">
              <col :style="`width:${tableUi.payments.widths.amount}px`">
              <col :style="`width:${tableUi.payments.widths.description}px`">
              <col :style="`width:${tableUi.payments.widths.date}px`">
              <col :style="`width:${tableUi.payments.widths.actions}px`">
            </colgroup>
            <thead>
              <tr>
                <th class="sortable" @click="toggleTableSort('payments', 'ID')">ID <span class="sort-indicator" x-text="tableSortIndicator('payments', 'ID')"></span><span class="col-resize-handle" @mousedown.stop.prevent="startColumnResize('payments', 'id', $event)"></span></th>
                <th class="sortable" @click="toggleTableSort('payments', 'PayerName')">Payer <span class="sort-indicator" x-text="tableSortIndicator('payments', 'PayerName')"></span><span class="col-resize-handle" @mousedown.stop.prevent="startColumnResize('payments', 'payer', $event)"></span></th>
                <th class="sortable" @click="toggleTableSort('payments', 'Owers')">Owers <span class="sort-indicator" x-text="tableSortIndicator('payments', 'Owers')"></span><span class="col-resize-handle" @mousedown.stop.prevent="startColumnResize('payments', 'owers', $event)"></span></th>
                <th class="amount-header sortable" @click="toggleTableSort('payments', 'Amount')">Amount <span class="sort-indicator" x-text="tableSortIndicator('payments', 'Amount')"></span><span class="col-resize-handle" @mousedown.stop.prevent="startColumnResize('payments', 'amount', $event)"></span></th>
                <th class="sortable" @click="toggleTableSort('payments', 'Description')">Description <span class="sort-indicator" x-text="tableSortIndicator('payments', 'Description')"></span><span class="col-resize-handle" @mousedown.stop.prevent="startColumnResize('payments', 'description', $event)"></span></th>
                <th class="sortable" @click="toggleTableSort('payments', 'Date')">Date <span class="sort-indicator" x-text="tableSortIndicator('payments', 'Date')"></span><span class="col-resize-handle" @mousedown.stop.prevent="startColumnResize('payments', 'date', $event)"></span></th>
                <th><span class="col-resize-handle" @mousedown.stop.prevent="startColumnResize('payments', 'actions', $event)"></span></th>
              </tr>
            </thead>
            <tbody>
              <template x-if="payments.length === 0">
                <tr class="empty-row"><td colspan="7">No payments yet</td></tr>
              </template>
              <template x-for="p in getTablePageRows('payments', payments)" :key="p.ID">
                <tr>
                  <td class="cell-id" x-text="p.ID"></td>
                  <td class="cell-id" x-text="p.PayerName"></td>
                  <td class="cell-id" x-text="p.Owers"></td>
                  <td class="payment-amount-cell">
                    <template x-if="!isShowingPaymentRate(p.ID)">
                      <button type="button" class="amount-btn payment-amount-group" @click="togglePaymentAmountView(p.ID)">
                        <template x-if="isCrossCurrencyPayment(p)">
                          <span class="payment-amount-group">
                            <span class="cell-amount" x-text="getFormattedAmount(p.Amount, p.FromExchangeRate)"></span>
                            <span class="balance-arrow">→</span>
                            <span class="cell-amount" x-text="getFormattedAmount(Number(p.Amount) * Number(p.ExchangeRate), p.ToExchangeRate)"></span>
                          </span>
                        </template>
                        <template x-if="!isCrossCurrencyPayment(p)">
                          <span class="cell-amount" x-text="getFormattedAmount(p.Amount, p.FromExchangeRate)"></span>
                        </template>
                      </button>
                    </template>
                    <template x-if="isShowingPaymentRate(p.ID)">
                      <button type="button" class="amount-btn payment-amount-group" @click="togglePaymentAmountView(p.ID)">
                        <span class="cell-amount" x-text="getPaymentRateFromDisplay(p)"></span>
                        <span class="balance-arrow">→</span>
                        <span class="cell-amount" x-text="getPaymentRateToDisplay(p)"></span>
                      </button>
                    </template>
                  </td>
                  <td x-text="p.Description"></td>
                  <td class="cell-date" x-text="p.Date"></td>
                  <td><button class="btn btn-danger" @click="deletePayment(p.ID)">Remove</button></td>
                </tr>
              </template>
            </tbody>
          </table>
        </div>

        <div class="table-pager" x-show="getTablePageCount('payments', payments.length) > 1">
          <button
            type="button"
            class="btn btn-ghost"
            @click="viewTablePage('payments', tableUi.payments.page - 1, payments.length)"
            :disabled="tableUi.payments.page === 0"
          >Prev</button>
          <div class="table-pager-info">
            Page <span x-text="tableUi.payments.page + 1"></span> of
            <span x-text="getTablePageCount('payments', payments.length)"></span> | Showing
            <span x-text="getTablePageStart('payments', payments.length)"></span> to
            <span x-text="getTablePageEnd('payments', payments.length)"></span>
          </div>
          <button
            type="button"
            class="btn btn-ghost"
            @click="viewTablePage('payments', tableUi.payments.page + 1, payments.length)"
            :disabled="tableUi.payments.page >= getTablePageCount('payments', payments.length) - 1"
          >Next</button>
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
            <colgroup>
              <col :style="`width:${tableUi.balances.widths.who}px`">
              <col :style="`width:${tableUi.balances.widths.amount}px`">
            </colgroup>
            <thead>
              <tr>
                <th class="sortable" @click="toggleTableSort('balances', 'FromUser')">Who owes whom <span class="sort-indicator" x-text="tableSortIndicator('balances', 'FromUser')"></span><span class="col-resize-handle" @mousedown.stop.prevent="startColumnResize('balances', 'who', $event)"></span></th>
                <th class="sortable" @click="toggleTableSort('balances', 'Amount')">Amount <span class="sort-indicator" x-text="tableSortIndicator('balances', 'Amount')"></span><span class="col-resize-handle" @mousedown.stop.prevent="startColumnResize('balances', 'amount', $event)"></span></th>
              </tr>
            </thead>
            <tbody>
              <template x-if="balances.length === 0">
                <tr class="empty-row"><td colspan="2">All settled up</td></tr>
              </template>
              <template x-for="b in getTablePageRows('balances', balances)" :key="`${b.FromUser}-${b.ToUser}-${b.Currency || ''}`">
                <tr>
                  <td>
                    <span class="balance-from" x-text="b.FromUser"></span>
                    <span class="balance-arrow">→</span>
                    <span class="balance-to" x-text="b.ToUser"></span>
                  </td>
                  <td><span class="balance-amount" x-text="getBalanceAmountDisplay(b)"></span></td>
                </tr>
              </template>
            </tbody>
          </table>
          <div class="table-pager" x-show="getTablePageCount('balances', balances.length) > 1">
            <button
              type="button"
              class="btn btn-ghost"
              @click="viewTablePage('balances', tableUi.balances.page - 1, balances.length)"
              :disabled="tableUi.balances.page === 0"
            >Prev</button>
            <div class="table-pager-info">
              Page <span x-text="tableUi.balances.page + 1"></span> of
              <span x-text="getTablePageCount('balances', balances.length)"></span> | Showing
              <span x-text="getTablePageStart('balances', balances.length)"></span> to
              <span x-text="getTablePageEnd('balances', balances.length)"></span>
            </div>
            <button
              type="button"
              class="btn btn-ghost"
              @click="viewTablePage('balances', tableUi.balances.page + 1, balances.length)"
              :disabled="tableUi.balances.page >= getTablePageCount('balances', balances.length) - 1"
            >Next</button>
          </div>
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
      const apiBase = "<?= h(rtrim($apiBase, "/")) ?>";
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
        exchangeRates: ["USD", "EUR", "JPY"],
        payments: [],
        balances: [],
        balancesRaw: "[]",
        balancesView: "table",
        payerOpen: false,
        owersOpen: false,
        fromExchangeRateOpen: false,
        toExchangeRateOpen: false,
        newUserName: "",
        newPayment: { amount: "", payer: "", description: "", fromExchangeRate: "USD", toExchangeRate: "USD", owers: [] },
        modal: { open: false, message: "", onConfirm: null },
        error: "",
        notice: "",
        activePaymentRateId: null,
        tableUi: {
          users: {
            page: 0,
            size: 10,
            sort: { key: "CreatedAt", dir: "desc" },
            widths: { id: 72, name: 180, created: 220, actions: 96 },
          },
          payments: {
            page: 0,
            size: 10,
            sort: { key: "Date", dir: "desc" },
            widths: { id: 64, payer: 110, owers: 140, amount: 210, description: 220, date: 176, actions: 96 },
          },
          balances: {
            page: 0,
            size: 10,
            sort: { key: "FromUser", dir: "asc" },
            widths: { who: 360, amount: 140 },
          },
        },
        resizeSensitivity: 1.35,

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

        getTablePageCount(tableName, totalCount) {
          const state = this.tableUi[tableName];
          if (!state) return 1;
          return Math.max(1, Math.ceil(totalCount / state.size));
        },

        getTablePageStart(tableName, totalCount) {
          const state = this.tableUi[tableName];
          if (!state || totalCount === 0) return 0;
          return state.page * state.size + 1;
        },

        getTablePageEnd(tableName, totalCount) {
          const state = this.tableUi[tableName];
          if (!state || totalCount === 0) return 0;
          return Math.min((state.page + 1) * state.size, totalCount);
        },

        viewTablePage(tableName, pageNumber, totalCount) {
          const state = this.tableUi[tableName];
          if (!state) return;
          const maxPage = Math.max(0, this.getTablePageCount(tableName, totalCount) - 1);
          state.page = Math.min(Math.max(pageNumber, 0), maxPage);
        },

        sortValue(tableName, row, key) {
          if (tableName === "payments" && key === "Owers") {
            const owers = row?.Owers;
            return Array.isArray(owers) ? owers.join(", ") : String(owers || "");
          }
          return row?.[key];
        },

        getSortedRows(tableName, rows) {
          const state = this.tableUi[tableName];
          if (!state) return rows;
          const arr = [...rows];
          const { key, dir } = state.sort;
          const direction = dir === "asc" ? 1 : -1;
          return arr.sort((a, b) => {
            const av = this.sortValue(tableName, a, key);
            const bv = this.sortValue(tableName, b, key);
            if (key === "Amount" || key === "ID") {
              return (Number(av) - Number(bv)) * direction;
            }
            if (key === "Date" || key === "CreatedAt") {
              return (new Date(av || 0).getTime() - new Date(bv || 0).getTime()) * direction;
            }
            return String(av || "").localeCompare(String(bv || "")) * direction;
          });
        },

        getTablePageRows(tableName, rows) {
          const state = this.tableUi[tableName];
          if (!state) return rows;
          const sorted = this.getSortedRows(tableName, rows);
          const start = state.page * state.size;
          return sorted.slice(start, start + state.size);
        },

        toggleTableSort(tableName, key) {
          const state = this.tableUi[tableName];
          if (!state) return;
          if (state.sort.key === key) {
            state.sort.dir = state.sort.dir === "asc" ? "desc" : "asc";
          } else {
            state.sort.key = key;
            state.sort.dir = (key === "Date" || key === "CreatedAt") ? "desc" : "asc";
          }
          state.page = 0;
        },

        tableSortIndicator(tableName, key) {
          const state = this.tableUi[tableName];
          if (!state || state.sort.key !== key) return "";
          return state.sort.dir === "asc" ? "↑" : "↓";
        },

        startColumnResize(tableName, widthKey, event) {
          const state = this.tableUi[tableName];
          if (!state || !state.widths || state.widths[widthKey] == null) return;
          const startX = event.clientX;
          const startWidth = state.widths[widthKey];
          const minWidth = widthKey === "actions" ? 72 : 60;
          const onMove = (moveEvent) => {
            const delta = (moveEvent.clientX - startX) * this.resizeSensitivity;
            state.widths[widthKey] = Math.max(minWidth, Math.round(startWidth + delta));
          };
          const onUp = () => {
            window.removeEventListener("mousemove", onMove);
            window.removeEventListener("mouseup", onUp);
            document.body.style.cursor = "";
            document.body.style.userSelect = "";
          };
          window.addEventListener("mousemove", onMove);
          window.addEventListener("mouseup", onUp);
          document.body.style.cursor = "col-resize";
          document.body.style.userSelect = "none";
        },

        clampAllTablePages() {
          const totals = {
            users: this.users.length,
            payments: this.payments.length,
            balances: this.balances.length,
          };
          Object.keys(this.tableUi).forEach((tableName) => {
            const state = this.tableUi[tableName];
            const maxPage = Math.max(0, this.getTablePageCount(tableName, totals[tableName] || 0) - 1);
            if (state.page > maxPage) state.page = maxPage;
          });
        },

        getFormattedAmount(amount, exchangeRate) {
          const numericAmount = Number(amount);
          if (!Number.isFinite(numericAmount)) return amount;

          const currency = (exchangeRate || "").toUpperCase();
          if (!currency) return numericAmount.toLocaleString();

          try {
            return new Intl.NumberFormat(undefined, {
              style: "currency",
              currency,
              currencyDisplay: "symbol",
            }).format(numericAmount);
          } catch (_) {
            return `${currency} ${numericAmount.toLocaleString()}`;
          }
        },

        getBalanceAmountDisplay(balance) {
          if (!balance) return "";
          return this.getFormattedAmount(balance.Amount, balance.Currency);
        },

        isCrossCurrencyPayment(payment) {
          if (!payment) return false;
          const from = (payment.FromExchangeRate || "").toUpperCase();
          const to = (payment.ToExchangeRate || "").toUpperCase();
          const rate = Number(payment.ExchangeRate);
          return Boolean(from && to && from !== to && Number.isFinite(rate));
        },

        togglePaymentAmountView(paymentId) {
          if (!paymentId) return;
          this.activePaymentRateId = this.activePaymentRateId === paymentId ? null : paymentId;
        },

        isShowingPaymentRate(paymentId) {
          return this.activePaymentRateId === paymentId;
        },

        getPaymentRateFromDisplay(payment) {
          if (!payment) return "";
          const from = (payment.FromExchangeRate || "").toUpperCase();
          if (!from) return "No exchange-rate data";
          return `1 ${from}`;
        },

        getPaymentRateToDisplay(payment) {
          if (!payment) return "";
          const to = (payment.ToExchangeRate || "").toUpperCase();
          const rate = Number(payment.ExchangeRate);
          if (!to || !Number.isFinite(rate)) return "No exchange-rate data";
          return `${rate.toLocaleString(undefined, { maximumFractionDigits: 6 })} ${to}`;
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
            this.clampAllTablePages();
            this.activePaymentRateId = null;
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
          if (!this.newPayment.payer) {
            this.flashError("Please select a payer");
            return;
          }
          const owers = Array.isArray(this.newPayment.owers)
            ? [...new Set(this.newPayment.owers.filter((name) => Boolean(name) && name !== this.newPayment.payer))]
            : [];
          try {
            await req("/api/v1/payment", {
              method: "POST",
              body: JSON.stringify({
                amount:      Number(this.newPayment.amount),
                payer:       this.newPayment.payer,
                description: this.newPayment.description,
                fromExchangeRate: this.newPayment.fromExchangeRate,
                toExchangeRate: this.newPayment.toExchangeRate,
                owers,
              }),
            });
            this.newPayment = { amount: "", payer: "", description: "", fromExchangeRate: "USD", toExchangeRate: "USD", owers: [] };
            this.payerOpen = false;
            this.owersOpen = false;
            this.fromExchangeRateOpen = false;
            this.toExchangeRateOpen = false;
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
