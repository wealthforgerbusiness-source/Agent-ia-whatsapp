package main

const loginHTML = `<!DOCTYPE html>
<html lang="fr">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Connexion - Agent WhatsApp</title>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background: linear-gradient(135deg, #075E54, #128C7E);
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
}
.card {
  background: white;
  border-radius: 16px;
  padding: 32px 24px;
  max-width: 360px;
  width: 100%;
  box-shadow: 0 10px 40px rgba(0,0,0,0.2);
}
h1 { font-size: 20px; margin-bottom: 24px; color: #075E54; text-align: center; }
input {
  width: 100%;
  padding: 14px;
  border: 1px solid #ddd;
  border-radius: 10px;
  font-size: 16px;
  margin-bottom: 16px;
}
button {
  width: 100%;
  padding: 14px;
  background: #075E54;
  color: white;
  border: none;
  border-radius: 10px;
  font-size: 16px;
  font-weight: 600;
  cursor: pointer;
}
button:hover { background: #064c44; }
.error {
  display: none;
  color: #d32f2f;
  font-size: 14px;
  margin-bottom: 12px;
  text-align: center;
}
</style>
</head>
<body>
<div class="card">
  <h1>🔒 Agent WhatsApp</h1>
  <div class="error">Mot de passe incorrect</div>
  <form method="POST" action="/login">
    <input type="password" name="password" placeholder="Mot de passe" required autofocus>
    <button type="submit">Se connecter</button>
  </form>
</div>
</body>
</html>`

const dashboardHTML = `<!DOCTYPE html>
<html lang="fr">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>Agent WhatsApp - Dashboard</title>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body {
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  background: #f0f2f5;
  padding: 16px;
  padding-bottom: 60px;
}
.container { max-width: 480px; margin: 0 auto; }
header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}
h1 { font-size: 18px; color: #075E54; }
a.logout { color: #999; font-size: 13px; text-decoration: none; }
.card {
  background: white;
  border-radius: 14px;
  padding: 20px;
  margin-bottom: 16px;
  box-shadow: 0 2px 8px rgba(0,0,0,0.06);
}
.card-title {
  font-size: 13px;
  color: #888;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  margin-bottom: 10px;
}
.row {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.status-dot {
  width: 10px;
  height: 10px;
  border-radius: 50%;
  display: inline-block;
  margin-right: 8px;
}
.dot-green { background: #25D366; }
.dot-red { background: #d32f2f; }
.dot-orange { background: #f5a623; }
.switch {
  position: relative;
  display: inline-block;
  width: 50px;
  height: 28px;
}
.switch input { opacity: 0; width: 0; height: 0; }
.slider {
  position: absolute;
  cursor: pointer;
  top: 0; left: 0; right: 0; bottom: 0;
  background-color: #ccc;
  transition: 0.3s;
  border-radius: 28px;
}
.slider:before {
  position: absolute;
  content: "";
  height: 22px; width: 22px;
  left: 3px; bottom: 3px;
  background-color: white;
  transition: 0.3s;
  border-radius: 50%;
}
input:checked + .slider { background-color: #25D366; }
input:checked + .slider:before { transform: translateX(22px); }
.big-number { font-size: 28px; font-weight: 700; color: #075E54; }
.sub-label { font-size: 13px; color: #888; margin-top: 4px; }
.progress-bar {
  width: 100%;
  height: 8px;
  background: #eee;
  border-radius: 4px;
  overflow: hidden;
  margin-top: 10px;
}
.progress-fill {
  height: 100%;
  background: #25D366;
  transition: width 0.3s;
}
.progress-fill.warning { background: #f5a623; }
.progress-fill.danger { background: #d32f2f; }
.qr-box { text-align: center; }
.qr-box img { max-width: 250px; width: 100%; border-radius: 8px; margin-top: 10px; }
button.reset-btn {
  width: 100%;
  padding: 12px;
  background: #f0f2f5;
  color: #333;
  border: 1px solid #ddd;
  border-radius: 10px;
  font-size: 14px;
  cursor: pointer;
  margin-top: 8px;
}
.badge {
  display: inline-block;
  padding: 3px 10px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 600;
}
.badge-green { background: #e6f8ee; color: #128C7E; }
.badge-red { background: #fdecea; color: #d32f2f; }
</style>
</head>
<body>
<div class="container">
  <header>
    <h1>🤖 Agent WhatsApp</h1>
    <a class="logout" href="/logout">Déconnexion</a>
  </header>

  <div class="card">
    <div class="card-title">Statut du bot</div>
    <div class="row">
      <div>
        <span id="op-badge" class="badge badge-red">Chargement...</span>
      </div>
      <label class="switch">
        <input type="checkbox" id="toggle-bot">
        <span class="slider"></span>
      </label>
    </div>
  </div>

  <div class="card">
    <div class="card-title">Connexion WhatsApp</div>
    <div class="row">
      <span><span id="wa-dot" class="status-dot dot-red"></span><span id="wa-status">Chargement...</span></span>
    </div>
    <div id="qr-container" class="qr-box" style="display:none;">
      <p style="font-size:13px; color:#888; margin-top:10px;">Scanne ce QR code avec WhatsApp pour connecter le bot</p>
      <img id="qr-img" src="">
    </div>
  </div>

  <div class="card">
    <div class="card-title">Abonnement</div>
    <div class="big-number" id="jours-restants">--</div>
    <div class="sub-label">jours restants avant expiration</div>
    <button class="reset-btn" id="reset-btn">Renouveler l'abonnement (30 jours)</button>
  </div>

  <div class="card">
    <div class="card-title">Utilisation Tokens (IA)</div>
    <div class="row">
      <span id="tokens-used">--</span>
      <span id="tokens-limit" style="color:#888;">/ --</span>
    </div>
    <div class="progress-bar"><div class="progress-fill" id="tokens-bar" style="width:0%"></div></div>
    <div class="sub-label" id="tokens-remaining"></div>
  </div>
</div>

<script>
let toggling = false;

async function fetchStatus() {
  try {
    const res = await fetch('/api/status');
    const data = await res.json();
    render(data);
  } catch (e) {
    console.error(e);
  }
}

function render(data) {
  const opBadge = document.getElementById('op-badge');
  if (data.bot_operationnel) {
    opBadge.textContent = 'Actif et opérationnel';
    opBadge.className = 'badge badge-green';
  } else if (!data.bot_actif) {
    opBadge.textContent = 'Désactivé manuellement';
    opBadge.className = 'badge badge-red';
  } else if (data.abonnement_expire) {
    opBadge.textContent = 'Abonnement expiré';
    opBadge.className = 'badge badge-red';
  } else if (data.limite_atteinte) {
    opBadge.textContent = 'Limite tokens atteinte';
    opBadge.className = 'badge badge-red';
  } else {
    opBadge.textContent = 'Inactif';
    opBadge.className = 'badge badge-red';
  }

  if (!toggling) {
    document.getElementById('toggle-bot').checked = data.bot_actif;
  }

  const waDot = document.getElementById('wa-dot');
  const waStatus = document.getElementById('wa-status');
  const qrContainer = document.getElementById('qr-container');
  const qrImg = document.getElementById('qr-img');

  if (data.whatsapp_status === 'connecte') {
    waDot.className = 'status-dot dot-green';
    waStatus.textContent = 'Connecté';
    qrContainer.style.display = 'none';
  } else if (data.whatsapp_status === 'attente_scan_qr') {
    waDot.className = 'status-dot dot-orange';
    waStatus.textContent = 'En attente de scan';
    if (data.qr_code) {
      qrContainer.style.display = 'block';
      qrImg.src = 'data:image/png;base64,' + data.qr_code;
    }
  } else {
    waDot.className = 'status-dot dot-red';
    waStatus.textContent = 'Déconnecté';
    qrContainer.style.display = 'none';
  }

  document.getElementById('jours-restants').textContent = data.jours_restants;

  const tokensUsed = data.tokens_total.toLocaleString('fr-FR');
  const tokensLimit = data.tokens_limite_mensuelle.toLocaleString('fr-FR');
  document.getElementById('tokens-used').textContent = tokensUsed + ' tokens utilisés';
  document.getElementById('tokens-limit').textContent = '/ ' + tokensLimit;

  const pct = Math.min(100, (data.tokens_total / data.tokens_limite_mensuelle) * 100);
  const bar = document.getElementById('tokens-bar');
  bar.style.width = pct + '%';
  bar.className = 'progress-fill' + (pct > 90 ? ' danger' : pct > 70 ? ' warning' : '');

  document.getElementById('tokens-remaining').textContent =
    data.tokens_restants.toLocaleString('fr-FR') + ' tokens restants ce mois';
}

document.getElementById('toggle-bot').addEventListener('change', async (e) => {
  toggling = true;
  await fetch('/api/toggle', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ actif: e.target.checked })
  });
  toggling = false;
  fetchStatus();
});

document.getElementById('reset-btn').addEventListener('click', async () => {
  if (!confirm('Renouveler l\'abonnement pour 30 jours et remettre les tokens à zéro ?')) return;
  await fetch('/api/reset', { method: 'POST' });
  fetchStatus();
});

fetchStatus();
setInterval(fetchStatus, 4000);
</script>
</body>
</html>`
