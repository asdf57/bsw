<?php
$apiBase = getenv('API_BASE');
if ($apiBase === false || $apiBase === '') {
    $apiBase = 'http://app:8080';
}

function fetchJson(string $url): array {
    $ch = curl_init($url);
    curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
    curl_setopt($ch, CURLOPT_CONNECTTIMEOUT, 2);
    curl_setopt($ch, CURLOPT_TIMEOUT, 5);
    $body = curl_exec($ch);
    $status = curl_getinfo($ch, CURLINFO_HTTP_CODE);
    $err = curl_error($ch);
    curl_close($ch);

    if ($body === false || $status < 200 || $status >= 300) {
        return [
            'ok' => false,
            'status' => $status,
            'error' => $err !== '' ? $err : 'request failed',
            'data' => null,
        ];
    }

    $data = json_decode($body, true);
    if (!is_array($data)) {
        return [
            'ok' => false,
            'status' => $status,
            'error' => 'invalid json',
            'data' => null,
        ];
    }

    return [
        'ok' => true,
        'status' => $status,
        'error' => null,
        'data' => $data,
    ];
}

$usersResp = fetchJson(rtrim($apiBase, '/') . '/api/v1/user/all');
$paymentsResp = fetchJson(rtrim($apiBase, '/') . '/api/v1/payment/all');

function h($value): string {
    return htmlspecialchars((string)$value, ENT_QUOTES, 'UTF-8');
}
?>
<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Better Splitwise</title>
  <style>
    body { font-family: Arial, sans-serif; margin: 24px; color: #111; }
    h1 { margin-bottom: 8px; }
    section { margin-top: 24px; }
    table { border-collapse: collapse; width: 100%; max-width: 960px; }
    th, td { border: 1px solid #ddd; padding: 8px; text-align: left; }
    th { background: #f5f5f5; }
    .error { color: #b00020; }
    .muted { color: #666; }
  </style>
</head>
<body>
  <h1>BSW</h1>
  <div class="muted">API: <?php echo h($apiBase); ?></div>

  <section>
    <h2>Active Users</h2>
    <?php if (!$usersResp['ok']): ?>
      <div class="error">Failed to load users (<?php echo h($usersResp['status']); ?>): <?php echo h($usersResp['error']); ?></div>
    <?php else: ?>
      <?php $users = $usersResp['data']; ?>
      <?php if (count($users) === 0): ?>
        <div class="muted">No users found.</div>
      <?php else: ?>
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>Name</th>
              <th>Created</th>
            </tr>
          </thead>
          <tbody>
            <?php foreach ($users as $u): ?>
              <tr>
                <td><?php echo h($u['ID'] ?? $u['id'] ?? ''); ?></td>
                <td><?php echo h($u['Name'] ?? $u['name'] ?? ''); ?></td>
                <td><?php echo h($u['CreatedAt'] ?? $u['created_at'] ?? ''); ?></td>
              </tr>
            <?php endforeach; ?>
          </tbody>
        </table>
      <?php endif; ?>
    <?php endif; ?>
  </section>

  <section>
    <h2>Payments</h2>
    <?php if (!$paymentsResp['ok']): ?>
      <div class="error">Failed to load payments (<?php echo h($paymentsResp['status']); ?>): <?php echo h($paymentsResp['error']); ?></div>
    <?php else: ?>
      <?php $payments = $paymentsResp['data']; ?>
      <?php if (count($payments) === 0): ?>
        <div class="muted">No payments found.</div>
      <?php else: ?>
        <table>
          <thead>
            <tr>
              <th>ID</th>
              <th>Amount</th>
              <th>Description</th>
              <th>Payer ID</th>
              <th>Date</th>
            </tr>
          </thead>
          <tbody>
            <?php foreach ($payments as $p): ?>
              <tr>
                <td><?php echo h($p['ID'] ?? $p['id'] ?? ''); ?></td>
                <td><?php echo h($p['Amount'] ?? $p['amount'] ?? ''); ?></td>
                <td><?php echo h($p['Description'] ?? $p['description'] ?? ''); ?></td>
                <td><?php echo h($p['PayerID'] ?? $p['payer_id'] ?? ''); ?></td>
                <td><?php echo h($p['Date'] ?? $p['date'] ?? ''); ?></td>
              </tr>
            <?php endforeach; ?>
          </tbody>
        </table>
      <?php endif; ?>
    <?php endif; ?>
  </section>
</body>
</html>
