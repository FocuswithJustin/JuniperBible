// Capsules page functionality

// Confirmation dialog handler for delete forms
function confirmDeleteCapsule(form) {
  const capsuleNameInput = form.querySelector('.capsule-name');
  const capsuleName = capsuleNameInput ? capsuleNameInput.value : 'this capsule';
  return confirm('Delete ' + capsuleName + '? This cannot be undone.');
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

// Initialize counts on load
document.addEventListener('DOMContentLoaded', function() {
  updateGenerateCount();
  updateExportCount();
});
