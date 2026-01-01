// internal/auth/templates.go
package auth

const setupTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Connect to Dub</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg: #FAFAF9;
            --bg-card: #FFFFFF;
            --bg-input: #F5F5F4;
            --bg-hint: #FFF7ED;
            --border: #E7E5E4;
            --border-focus: #404040;
            --text: #404040;
            --text-secondary: #737373;
            --text-muted: #A3A3A3;
            --accent: #F97316;
            --accent-dark: #EA580C;
            --accent-light: #FFF7ED;
            --success: #16A34A;
            --success-light: #DCFCE7;
            --error: #DC2626;
            --error-light: #FEE2E2;
        }

        * { margin: 0; padding: 0; box-sizing: border-box; }

        html { height: 100%; }

        body {
            font-family: 'Plus Jakarta Sans', -apple-system, sans-serif;
            background: var(--bg);
            color: var(--text);
            min-height: 100%;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            padding: 2rem 1.5rem 3rem;
        }

        .container {
            width: 100%;
            max-width: 400px;
        }

        /* Logo section */
        .header {
            text-align: center;
            margin-bottom: 1.5rem;
        }

        .logo {
            display: inline-flex;
            align-items: center;
            justify-content: center;
            width: 48px;
            height: 48px;
            background: var(--accent);
            border-radius: 12px;
            margin-bottom: 1rem;
            box-shadow: 0 4px 14px rgba(249, 115, 22, 0.3);
        }

        .logo svg {
            width: 24px;
            height: 24px;
            color: white;
        }

        .badge {
            display: inline-flex;
            align-items: center;
            gap: 0.375rem;
            background: var(--accent-light);
            color: var(--text-secondary);
            font-size: 0.6875rem;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.05em;
            padding: 0.375rem 0.75rem;
            border-radius: 100px;
            margin-bottom: 0.75rem;
        }

        .badge svg {
            width: 12px;
            height: 12px;
        }

        h1 {
            font-size: 1.375rem;
            font-weight: 700;
            letter-spacing: -0.02em;
            margin-bottom: 0.25rem;
        }

        .subtitle {
            color: var(--text-secondary);
            font-size: 0.875rem;
        }

        /* Help card */
        .help-card {
            background: var(--bg-hint);
            border-radius: 12px;
            padding: 0.875rem;
            margin-bottom: 1rem;
        }

        .help-header {
            display: flex;
            align-items: center;
            gap: 0.5rem;
            font-size: 0.75rem;
            font-weight: 500;
            color: var(--text-secondary);
            margin-bottom: 0.625rem;
        }

        .help-header svg {
            width: 14px;
            height: 14px;
            color: var(--text-muted);
        }

        .help-link {
            display: flex;
            align-items: center;
            gap: 0.625rem;
            padding: 0.625rem 0.75rem;
            background: var(--bg-card);
            border: 1px solid var(--border);
            border-radius: 10px;
            text-decoration: none;
            color: var(--text);
            transition: all 0.15s ease;
        }

        .help-link:hover {
            border-color: var(--accent);
            box-shadow: 0 0 0 2px var(--accent-light);
        }

        .help-icon {
            width: 32px;
            height: 32px;
            background: var(--accent);
            border-radius: 8px;
            display: flex;
            align-items: center;
            justify-content: center;
            flex-shrink: 0;
        }

        .help-icon svg {
            width: 16px;
            height: 16px;
            color: white;
        }

        .help-text {
            flex: 1;
            min-width: 0;
        }

        .help-title {
            font-weight: 600;
            font-size: 0.8125rem;
        }

        .help-path {
            font-size: 0.6875rem;
            color: var(--text-muted);
        }

        .help-arrow {
            color: var(--text-muted);
            flex-shrink: 0;
        }

        /* Form card */
        .form-card {
            background: var(--bg-card);
            border: 1px solid var(--border);
            border-radius: 14px;
            padding: 1.25rem;
            box-shadow: 0 1px 3px rgba(0, 0, 0, 0.04), 0 4px 12px rgba(0, 0, 0, 0.02);
        }

        .form-group {
            margin-bottom: 1rem;
        }

        .form-group:last-of-type {
            margin-bottom: 0;
        }

        .label-row {
            display: flex;
            align-items: center;
            justify-content: space-between;
            margin-bottom: 0.375rem;
        }

        label {
            font-size: 0.8125rem;
            font-weight: 600;
            color: var(--text);
        }

        .required-badge {
            font-size: 0.625rem;
            font-weight: 600;
            text-transform: uppercase;
            letter-spacing: 0.04em;
            padding: 0.125rem 0.4rem;
            border-radius: 4px;
            background: var(--accent-light);
            color: var(--text-secondary);
        }

        .input-wrapper {
            position: relative;
        }

        input {
            width: 100%;
            padding: 0.625rem 0.75rem;
            font-family: inherit;
            font-size: 0.8125rem;
            background: var(--bg-input);
            border: 1.5px solid transparent;
            border-radius: 8px;
            color: var(--text);
            transition: all 0.15s ease;
        }

        input::placeholder {
            color: var(--text-muted);
        }

        input:focus {
            outline: none;
            background: var(--bg-card);
            border-color: var(--accent);
            box-shadow: 0 0 0 3px rgba(249, 115, 22, 0.12);
        }

        input.mono {
            font-family: 'JetBrains Mono', monospace;
            font-size: 0.75rem;
            letter-spacing: -0.01em;
        }

        input.error {
            border-color: var(--error);
            background: var(--error-light);
        }

        .input-hint {
            font-size: 0.6875rem;
            color: var(--text-muted);
            margin-top: 0.25rem;
        }

        /* Password toggle */
        .password-toggle {
            position: absolute;
            right: 0.5rem;
            top: 50%;
            transform: translateY(-50%);
            background: none;
            border: none;
            color: var(--text-muted);
            cursor: pointer;
            padding: 0.25rem;
            border-radius: 4px;
            display: flex;
            align-items: center;
            justify-content: center;
        }

        .password-toggle:hover {
            color: var(--text-secondary);
        }

        .password-toggle svg {
            width: 16px;
            height: 16px;
        }

        /* Buttons */
        .btn-group {
            display: flex;
            gap: 0.5rem;
            margin-top: 1.25rem;
        }

        button {
            flex: 1;
            padding: 0.625rem 1rem;
            font-family: inherit;
            font-size: 0.8125rem;
            font-weight: 600;
            border-radius: 8px;
            cursor: pointer;
            transition: all 0.15s ease;
            border: none;
        }

        .btn-secondary {
            background: var(--bg-input);
            color: var(--text-secondary);
            border: 1px solid var(--border);
        }

        .btn-secondary:hover:not(:disabled) {
            background: var(--border);
            color: var(--text);
        }

        .btn-primary {
            background: var(--accent);
            color: white;
        }

        .btn-primary:hover:not(:disabled) {
            background: var(--accent-dark);
            transform: translateY(-1px);
            box-shadow: 0 4px 12px rgba(249, 115, 22, 0.3);
        }

        button:disabled {
            opacity: 0.5;
            cursor: not-allowed;
        }

        /* Toast status */
        .status {
            position: fixed;
            bottom: 1.5rem;
            left: 50%;
            transform: translateX(-50%) translateY(10px);
            padding: 0.625rem 1rem;
            border-radius: 10px;
            font-size: 0.8125rem;
            font-weight: 500;
            display: flex;
            align-items: center;
            gap: 0.5rem;
            opacity: 0;
            visibility: hidden;
            transition: all 0.25s cubic-bezier(0.4, 0, 0.2, 1);
            box-shadow: 0 4px 20px rgba(0, 0, 0, 0.12);
            z-index: 100;
            white-space: nowrap;
        }

        .status.show {
            opacity: 1;
            visibility: visible;
            transform: translateX(-50%) translateY(0);
        }

        .status.loading {
            background: var(--accent-light);
            color: var(--text);
        }

        .status.success {
            background: var(--success-light);
            color: var(--success);
        }

        .status.error {
            background: var(--error-light);
            color: var(--error);
        }

        .spinner {
            width: 14px;
            height: 14px;
            border: 2px solid currentColor;
            border-top-color: transparent;
            border-radius: 50%;
            animation: spin 0.6s linear infinite;
        }

        @keyframes spin { to { transform: rotate(360deg); } }

        .status-icon {
            width: 14px;
            height: 14px;
            flex-shrink: 0;
        }

        /* Footer */
        .footer {
            margin-top: 1.5rem;
            display: flex;
            justify-content: center;
            gap: 1.5rem;
        }

        .footer a {
            display: inline-flex;
            align-items: center;
            gap: 0.375rem;
            color: var(--text-muted);
            font-size: 0.75rem;
            font-weight: 500;
            text-decoration: none;
            transition: color 0.15s ease;
        }

        .footer a:hover {
            color: var(--text);
        }

        .footer svg {
            width: 15px;
            height: 15px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <div class="logo">
                <svg width="28" height="28" viewBox="0 0 64 64" fill="none" xmlns="http://www.w3.org/2000/svg">
                    <path fill-rule="evenodd" clip-rule="evenodd" d="M32 64c17.673 0 32-14.327 32-32 0-11.844-6.435-22.186-16-27.719V48h-8v-2.14A15.9 15.9 0 0 1 32 48c-8.837 0-16-7.163-16-16s7.163-16 16-16c2.914 0 5.647.78 8 2.14V1.008A32 32 0 0 0 32 0C14.327 0 0 14.327 0 32s14.327 32 32 32" fill="#fff"/>
                </svg>
            </div>
            <div class="badge">
                <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5">
                    <rect x="2" y="2" width="12" height="12" rx="2"/>
                    <path d="M5 6L7 8L5 10M9 10H11" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
                CLI Setup
            </div>
            <h1>Connect to Dub</h1>
            <p class="subtitle">Enter your API credentials to get started</p>
        </div>

        <div class="help-card">
            <div class="help-header">
                <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5">
                    <circle cx="8" cy="8" r="6"/>
                    <path d="M8 7V11M8 5V5.01" stroke-linecap="round"/>
                </svg>
                Where to find your API key
            </div>
            <a href="https://app.dub.co/settings/tokens" target="_blank" class="help-link">
                <div class="help-icon">
                    <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5">
                        <path d="M9.5 4.5L11.5 2.5a2 2 0 0 1 2.83 2.83L12.5 7" stroke-linecap="round"/>
                        <path d="M6.5 11.5L4.5 13.5a2 2 0 0 1-2.83-2.83L3.5 9" stroke-linecap="round"/>
                        <path d="M10 6L6 10" stroke-linecap="round"/>
                    </svg>
                </div>
                <div class="help-text">
                    <div class="help-title">Dub Dashboard → API Tokens</div>
                    <div class="help-path">app.dub.co → Settings → Tokens</div>
                </div>
                <svg class="help-arrow" width="16" height="16" viewBox="0 0 16 16" fill="none">
                    <path d="M6 4L10 8L6 12" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"/>
                </svg>
            </a>
        </div>

        <div class="form-card">
            <form id="setupForm" autocomplete="off">
                <input type="hidden" name="csrf_token" id="csrf_token" value="{{.CSRFToken}}">
                <div class="form-group">
                    <div class="label-row">
                        <label for="workspace">Workspace Name</label>
                        <span class="required-badge">Required</span>
                    </div>
                    <input type="text" id="workspace" name="workspace" class="mono" placeholder="e.g., acme-prod" required autofocus>
                    <div class="input-hint">Name this config anything you like — useful if you manage multiple Dub workspaces</div>
                </div>

                <div class="form-group">
                    <div class="label-row">
                        <label for="api_key">API Key</label>
                        <span class="required-badge">Required</span>
                    </div>
                    <div class="input-wrapper">
                        <input type="password" id="api_key" name="api_key" class="mono" placeholder="dub_xxxxxxxxxxxxxxxx" required style="padding-right: 2rem;">
                        <button type="button" class="password-toggle" id="togglePassword" aria-label="Toggle visibility">
                            <svg id="eyeIcon" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5">
                                <path d="M2 8s2.5-4 6-4 6 4 6 4-2.5 4-6 4-6-4-6-4Z"/>
                                <circle cx="8" cy="8" r="1.5"/>
                            </svg>
                            <svg id="eyeOffIcon" style="display:none" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5">
                                <path d="M6.5 6.5a2 2 0 1 0 3 3M11 11.5A6.5 6.5 0 0 1 8 12c-3.5 0-6-4-6-4a10 10 0 0 1 2.5-3m2-1.5A5 5 0 0 1 8 4c3.5 0 6 4 6 4-.3.5-.8 1.2-1.5 2"/>
                                <path d="M2 2l12 12"/>
                            </svg>
                        </button>
                    </div>
                    <div class="input-hint">Your Dub API key starting with dub_</div>
                </div>

                <div class="btn-group">
                    <button type="button" id="testBtn" class="btn-secondary">Test</button>
                    <button type="submit" id="submitBtn" class="btn-primary">Save & Connect</button>
                </div>
            </form>
        </div>

        <div class="footer">
            <a href="https://github.com/salmonumbrella/dub-cli" target="_blank">
                <svg viewBox="0 0 16 16" fill="currentColor">
                    <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0016 8c0-4.42-3.58-8-8-8z"/>
                </svg>
                View on GitHub
            </a>
            <a href="https://dub.co/docs/api-reference" target="_blank">
                <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5">
                    <path d="M2 3h12v10H2zM5 3v10"/>
                    <path d="M8 6h4M8 8h4M8 10h2" stroke-linecap="round"/>
                </svg>
                API Docs
            </a>
        </div>
    </div>

    <div id="status" class="status"></div>

    <script>
        const form = document.getElementById('setupForm');
        const testBtn = document.getElementById('testBtn');
        const submitBtn = document.getElementById('submitBtn');
        const status = document.getElementById('status');
        const togglePassword = document.getElementById('togglePassword');
        const apiKeyInput = document.getElementById('api_key');
        const workspaceInput = document.getElementById('workspace');
        const eyeIcon = document.getElementById('eyeIcon');
        const eyeOffIcon = document.getElementById('eyeOffIcon');
        function getCsrfToken() {
            const el = document.getElementById('csrf_token');
            return el ? el.value : '';
        }

        let isBusy = false;

        // Clear error state on input
        [workspaceInput, apiKeyInput].forEach(el => {
            el.addEventListener('input', () => el.classList.remove('error'));
        });

        togglePassword.addEventListener('click', () => {
            const isPassword = apiKeyInput.type === 'password';
            apiKeyInput.type = isPassword ? 'text' : 'password';
            eyeIcon.style.display = isPassword ? 'none' : 'block';
            eyeOffIcon.style.display = isPassword ? 'block' : 'none';
        });

        function showStatus(type, message) {
            status.className = 'status show ' + type;
            if (type === 'loading') {
                status.innerHTML = '<div class="spinner"></div><span>' + message + '</span>';
            } else {
                const icon = type === 'success'
                    ? '<svg class="status-icon" viewBox="0 0 16 16" fill="none"><path d="M13 5L6.5 11.5L3 8" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/></svg>'
                    : '<svg class="status-icon" viewBox="0 0 16 16" fill="none"><path d="M12 4L4 12M4 4L12 12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/></svg>';
                status.innerHTML = icon + '<span>' + message + '</span>';
            }
        }

        function hideStatus() {
            status.className = 'status';
        }

        function validateForm() {
            let valid = true;
            const workspace = workspaceInput.value.trim();
            const apiKey = apiKeyInput.value.trim();

            if (!workspace) {
                workspaceInput.classList.add('error');
                valid = false;
            }
            if (!apiKey) {
                apiKeyInput.classList.add('error');
                valid = false;
            } else if (!apiKey.startsWith('dub_')) {
                apiKeyInput.classList.add('error');
                showStatus('error', 'API key must start with dub_');
                return false;
            }
            return valid;
        }

        testBtn.addEventListener('click', async () => {
            if (isBusy) return;
            hideStatus();

            const apiKey = apiKeyInput.value.trim();
            if (!apiKey) {
                apiKeyInput.classList.add('error');
                return;
            }
            if (!apiKey.startsWith('dub_')) {
                apiKeyInput.classList.add('error');
                showStatus('error', 'API key must start with dub_');
                return;
            }

            isBusy = true;
            testBtn.disabled = true;
            submitBtn.disabled = true;
            showStatus('loading', 'Testing connection...');

            try {
                const csrfVal = getCsrfToken();
                const params = new URLSearchParams();
                params.append('csrf_token', csrfVal);
                params.append('api_key', apiKey);

                const resp = await fetch('/validate', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/x-www-form-urlencoded'},
                    body: params.toString()
                });
                const data = await resp.json();

                if (resp.ok && data.status === 'valid') {
                    showStatus('success', 'Connection successful!');
                } else {
                    showStatus('error', data.error || 'Connection failed');
                }
            } catch (err) {
                showStatus('error', 'Request failed');
            } finally {
                testBtn.disabled = false;
                submitBtn.disabled = false;
                isBusy = false;
            }
        });

        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            if (isBusy) return;
            hideStatus();

            if (!validateForm()) return;

            isBusy = true;
            testBtn.disabled = true;
            submitBtn.disabled = true;
            showStatus('loading', 'Saving credentials...');

            try {
                const csrfVal = getCsrfToken();
                const params = new URLSearchParams();
                params.append('csrf_token', csrfVal);
                params.append('workspace', workspaceInput.value.trim());
                params.append('api_key', apiKeyInput.value.trim());

                const resp = await fetch('/submit', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/x-www-form-urlencoded'},
                    body: params.toString()
                });
                const data = await resp.json();

                if (resp.ok && data.redirect) {
                    showStatus('success', 'Saved! Redirecting...');
                    setTimeout(() => { window.location.href = data.redirect; }, 400);
                } else {
                    showStatus('error', data.error || 'Failed to save');
                    testBtn.disabled = false;
                    submitBtn.disabled = false;
                    isBusy = false;
                }
            } catch (err) {
                showStatus('error', 'Request failed');
                testBtn.disabled = false;
                submitBtn.disabled = false;
                isBusy = false;
            }
        });
    </script>
</body>
</html>`

const successTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Connected - Dub CLI</title>
    <link rel="preconnect" href="https://fonts.googleapis.com">
    <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
    <link href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;600;700&family=JetBrains+Mono:wght@400;500&display=swap" rel="stylesheet">
    <style>
        :root {
            --bg: #FAFAF9;
            --bg-card: #FFFFFF;
            --bg-terminal: #1C1917;
            --border: #E7E5E4;
            --text: #404040;
            --text-secondary: #737373;
            --text-muted: #A3A3A3;
            --accent: #F97316;
            --accent-dark: #EA580C;
            --accent-light: #FFF7ED;
            --success: #16A34A;
            --success-light: #DCFCE7;
        }

        * { margin: 0; padding: 0; box-sizing: border-box; }
        html { height: 100%; }

        body {
            font-family: 'Plus Jakarta Sans', -apple-system, sans-serif;
            background: var(--bg);
            color: var(--text);
            min-height: 100%;
            display: flex;
            flex-direction: column;
            align-items: center;
            justify-content: center;
            padding: 2rem 1.5rem 3rem;
        }

        .container {
            width: 100%;
            max-width: 400px;
            text-align: center;
        }

        /* Success icon */
        .success-icon {
            width: 56px;
            height: 56px;
            background: var(--success-light);
            border-radius: 50%;
            margin: 0 auto 1.25rem;
            display: flex;
            align-items: center;
            justify-content: center;
            animation: scaleIn 0.4s cubic-bezier(0.34, 1.56, 0.64, 1) forwards;
        }

        @keyframes scaleIn {
            from { transform: scale(0); opacity: 0; }
            to { transform: scale(1); opacity: 1; }
        }

        .success-icon svg {
            width: 28px;
            height: 28px;
            color: var(--success);
        }

        h1 {
            font-size: 1.375rem;
            font-weight: 700;
            letter-spacing: -0.02em;
            margin-bottom: 0.25rem;
            animation: fadeUp 0.5s ease 0.15s both;
        }

        .subtitle {
            color: var(--text-secondary);
            font-size: 0.875rem;
            margin-bottom: 1rem;
            animation: fadeUp 0.5s ease 0.25s both;
        }

        @keyframes fadeUp {
            from { opacity: 0; transform: translateY(8px); }
            to { opacity: 1; transform: translateY(0); }
        }

        /* Workspace badge */
        .workspace-badge {
            display: inline-flex;
            align-items: center;
            gap: 0.375rem;
            background: var(--accent-light);
            border: 1px solid rgba(249, 115, 22, 0.2);
            color: var(--accent-dark);
            font-family: 'JetBrains Mono', monospace;
            font-size: 0.75rem;
            font-weight: 500;
            padding: 0.375rem 0.75rem;
            border-radius: 100px;
            margin-bottom: 1.25rem;
            animation: fadeUp 0.5s ease 0.3s both;
        }

        .workspace-badge .dot {
            width: 6px;
            height: 6px;
            background: var(--success);
            border-radius: 50%;
            animation: pulse 2s ease-in-out infinite;
        }

        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.4; }
        }

        /* Terminal */
        .terminal {
            background: var(--bg-terminal);
            border-radius: 12px;
            overflow: hidden;
            text-align: left;
            animation: fadeUp 0.5s ease 0.35s both;
            box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
        }

        .terminal-bar {
            background: #0C0A09;
            padding: 0.625rem 0.75rem;
            display: flex;
            align-items: center;
            gap: 0.375rem;
        }

        .terminal-dot {
            width: 10px;
            height: 10px;
            border-radius: 50%;
        }

        .terminal-dot.red { background: #EF4444; }
        .terminal-dot.yellow { background: #F59E0B; }
        .terminal-dot.green { background: #22C55E; }

        .terminal-body {
            padding: 0.875rem 1rem;
        }

        .terminal-line {
            display: flex;
            align-items: center;
            gap: 0.5rem;
            font-family: 'JetBrains Mono', monospace;
            font-size: 0.75rem;
            line-height: 1.6;
            color: #D6D3D1;
        }

        .terminal-line + .terminal-line { margin-top: 0.25rem; }

        .terminal-prompt {
            color: var(--accent);
            user-select: none;
        }

        .terminal-cmd {
            color: #22C55E;
        }

        .terminal-output {
            color: #78716C;
            font-size: 0.6875rem;
            padding-left: 1rem;
            margin-top: 0.125rem;
            margin-bottom: 0.375rem;
        }

        .terminal-cursor {
            display: inline-block;
            width: 8px;
            height: 14px;
            background: var(--accent);
            animation: blink 1.2s step-end infinite;
            margin-left: 2px;
            vertical-align: middle;
        }

        @keyframes blink {
            0%, 50% { opacity: 1; }
            50.01%, 100% { opacity: 0; }
        }

        /* Message card */
        .message {
            margin-top: 1rem;
            padding: 1rem;
            background: var(--bg-card);
            border: 1px solid var(--border);
            border-radius: 12px;
            animation: fadeUp 0.5s ease 0.45s both;
        }

        .message-icon {
            display: inline-flex;
            align-items: center;
            justify-content: center;
            width: 32px;
            height: 32px;
            background: var(--accent);
            border-radius: 8px;
            margin-bottom: 0.625rem;
        }

        .message-icon svg {
            width: 16px;
            height: 16px;
            color: white;
        }

        .message-title {
            font-weight: 600;
            font-size: 0.875rem;
            margin-bottom: 0.25rem;
        }

        .message-text {
            font-size: 0.75rem;
            color: var(--text-secondary);
            line-height: 1.5;
        }

        .message-text code {
            font-family: 'JetBrains Mono', monospace;
            background: var(--accent-light);
            color: var(--accent-dark);
            padding: 0.125rem 0.375rem;
            border-radius: 4px;
            font-size: 0.6875rem;
        }

        /* Footer */
        .footer {
            margin-top: 1rem;
            font-size: 0.75rem;
            color: var(--text-muted);
            animation: fadeUp 0.5s ease 0.55s both;
        }

        .footer a {
            display: inline-flex;
            align-items: center;
            gap: 0.375rem;
            color: var(--text-muted);
            text-decoration: none;
            transition: color 0.15s ease;
            margin-top: 0.5rem;
        }

        .footer a:hover {
            color: var(--text-secondary);
        }

        .footer svg {
            width: 14px;
            height: 14px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="success-icon">
            <svg viewBox="0 0 28 28" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round">
                <path d="M6 14L12 20L22 8"/>
            </svg>
        </div>

        <h1>You're all set!</h1>
        <p class="subtitle">Dub CLI is connected and ready to use</p>

        {{if .Workspace}}
        <div class="workspace-badge">
            <span class="dot"></span>
            <span>{{.Workspace}}</span>
        </div>
        {{end}}

        <div class="terminal">
            <div class="terminal-bar">
                <span class="terminal-dot red"></span>
                <span class="terminal-dot yellow"></span>
                <span class="terminal-dot green"></span>
            </div>
            <div class="terminal-body">
                <div class="terminal-line">
                    <span class="terminal-prompt">$</span>
                    <span class="terminal-cmd">dub</span>
                    <span>links list</span>
                </div>
                <div class="terminal-output">Fetching your short links...</div>
                <div class="terminal-line">
                    <span class="terminal-prompt">$</span>
                    <span class="terminal-cmd">dub</span>
                    <span>links create https://example.com</span>
                </div>
                <div class="terminal-output">Creating short link...</div>
                <div class="terminal-line">
                    <span class="terminal-prompt">$</span>
                    <span class="terminal-cursor"></span>
                </div>
            </div>
        </div>

        <div class="message">
            <div class="message-icon">
                <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round">
                    <path d="M10 2L4 8l6 6"/>
                </svg>
            </div>
            <div class="message-title">Return to your terminal</div>
            <div class="message-text">
                You can close this window. Try <code>dub --help</code> to see all commands.
            </div>
        </div>

        <p class="footer">
            This window will close automatically
            <br>
            <a href="https://dub.co/docs/api-reference" target="_blank">
                <svg viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5">
                    <path d="M2 3h12v10H2zM5 3v10"/>
                    <path d="M8 6h4M8 8h4M8 10h2" stroke-linecap="round"/>
                </svg>
                API Documentation
            </a>
        </p>
    </div>

    <script>
        fetch('/complete?csrf={{.CSRFToken}}', { method: 'POST' }).catch(() => {});
    </script>
</body>
</html>`
