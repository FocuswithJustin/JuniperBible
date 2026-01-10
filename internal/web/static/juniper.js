// Juniper Tab JavaScript
// Handles capsules tab selection functions and repository refresh cooldown

// Custom source URL handling
function handleSourceChange(select) {
  const customContainer = document.getElementById('custom-url-container');
  const sourceUrl = document.getElementById('source-url');
  if (select.value === 'Custom') {
    customContainer.style.display = 'flex';
    customContainer.style.gap = '0.5rem';
    customContainer.style.alignItems = 'center';
    sourceUrl.style.display = 'none';
    document.getElementById('custom-url').focus();
  } else {
    customContainer.style.display = 'none';
    sourceUrl.style.display = '';
    select.form.submit();
  }
}

// Generate IR tab functions
function selectAllGenerate() {
  document.querySelectorAll('#generate-table input[name="source"]').forEach(cb => cb.checked = true);
  updateGenerateCount();
  updateSelectAllGenerateState();
}

function selectNoneGenerate() {
  document.querySelectorAll('#generate-table input[name="source"]').forEach(cb => cb.checked = false);
  updateGenerateCount();
  updateSelectAllGenerateState();
}

function toggleSelectAllGenerate(checkbox) {
  if (checkbox.checked) {
    selectAllGenerate();
  } else {
    selectNoneGenerate();
  }
}

function updateGenerateCount() {
  const checked = document.querySelectorAll('#generate-table input[name="source"]:checked').length;
  const el = document.getElementById('generate-selected-count');
  if (el) el.textContent = checked + ' selected';
}

function updateSelectAllGenerateState() {
  const all = document.querySelectorAll('#generate-table input[name="source"]');
  const checked = document.querySelectorAll('#generate-table input[name="source"]:checked');
  const selectAll = document.getElementById('select-all-generate');
  if (selectAll) selectAll.checked = all.length > 0 && all.length === checked.length;
}

// Export tab functions
function selectAllExport() {
  document.querySelectorAll('#export-table input[name="source"]').forEach(cb => cb.checked = true);
  updateExportCount();
  updateSelectAllExportState();
}

function selectNoneExport() {
  document.querySelectorAll('#export-table input[name="source"]').forEach(cb => cb.checked = false);
  updateExportCount();
  updateSelectAllExportState();
}

function toggleSelectAllExport(checkbox) {
  if (checkbox.checked) {
    selectAllExport();
  } else {
    selectNoneExport();
  }
}

function updateExportCount() {
  const checked = document.querySelectorAll('#export-table input[name="source"]:checked').length;
  const el = document.getElementById('export-selected-count');
  if (el) el.textContent = checked + ' selected';
}

function updateSelectAllExportState() {
  const all = document.querySelectorAll('#export-table input[name="source"]');
  const checked = document.querySelectorAll('#export-table input[name="source"]:checked');
  const selectAll = document.getElementById('select-all-export');
  if (selectAll) selectAll.checked = all.length > 0 && all.length === checked.length;
}

// Initialize refresh cooldown system
function initRefreshCooldown(selectedSource) {
  const COOLDOWN_KEY = 'repoman_refresh_cooldown_' + selectedSource;
  const COOLDOWN_SECONDS = 30;
  const btn = document.getElementById('load-btn');
  const btnText = document.getElementById('btn-text');
  const btnCountdown = document.getElementById('btn-countdown');
  const form = document.getElementById('load-form');

  if (!btn || !btnText || !btnCountdown || !form) return;

  function formatTime(seconds) {
    return seconds + 's';
  }

  function updateButtonState() {
    const cooldownEnd = localStorage.getItem(COOLDOWN_KEY);
    if (cooldownEnd) {
      const remaining = Math.ceil((parseInt(cooldownEnd) - Date.now()) / 1000);
      if (remaining > 0) {
        btn.disabled = true;
        btn.setAttribute('aria-busy', 'true');
        btnText.style.display = 'none';
        btnCountdown.style.display = 'inline';
        btnCountdown.textContent = 'Refresh available in ' + formatTime(remaining);
        return true;
      } else {
        localStorage.removeItem(COOLDOWN_KEY);
      }
    }
    btn.disabled = false;
    btn.removeAttribute('aria-busy');
    btnText.style.display = 'inline';
    btnCountdown.style.display = 'none';
    return false;
  }

  function startCooldown() {
    const cooldownEnd = Date.now() + (COOLDOWN_SECONDS * 1000);
    localStorage.setItem(COOLDOWN_KEY, cooldownEnd.toString());
  }

  // Check cooldown state on load
  if (updateButtonState()) {
    const interval = setInterval(function() {
      if (!updateButtonState()) {
        clearInterval(interval);
      }
    }, 1000);
  }

  // Start cooldown on form submit
  form.addEventListener('submit', function() {
    startCooldown();
  });
}

// Initialize on page load
document.addEventListener('DOMContentLoaded', function() {
  // Initialize custom source select
  const select = document.getElementById('source');
  if (select && select.value === 'Custom') {
    document.getElementById('custom-url-container').style.display = 'flex';
    document.getElementById('custom-url-container').style.gap = '0.5rem';
    document.getElementById('custom-url-container').style.alignItems = 'center';
    document.getElementById('source-url').style.display = 'none';
  }

  // Initialize counts on capsules tab
  const params = new URLSearchParams(window.location.search);
  if (params.get('tab') === 'capsules') {
    updateGenerateCount();
    updateExportCount();
  }
});
