const assert = require('node:assert/strict');
const { readFileSync } = require('node:fs');
const path = require('node:path');
const test = require('node:test');
const vm = require('node:vm');

// Exercise the template's actual scripts without a browser or extra dependencies.
function loadPartnerPage(isArchive = true) {
  const template = readFileSync(path.join(__dirname, '../templates/master/partner/index.html'), 'utf8');
  const state = { reloads: [], events: [], requests: [], dialogs: [], validation: [] };
  const table = {
    ajax: { reload: (...args) => state.reloads.push(args) },
    row: () => ({ data: () => state.row }),
  };
  const jquery = () => ({
    DataTable: (config) => {
      if (config) state.config = config;
      return table;
    },
    on: (event, selector, handler) => {
      if (selector === 'button[data-action="delete-permanent"]') state.clickDelete = handler;
    },
    closest: () => ({}),
  });
  const context = {
    $: jquery,
    console,
    document: { addEventListener: (event, callback) => callback() },
    window: { dispatchEvent: (event) => state.events.push(event) },
    CustomEvent: class { constructor(type, options) { this.type = type; this.detail = options.detail; } },
    fetch: async (...args) => {
      state.requests.push(args);
      if (state.fetchError) throw state.fetchError;
      return state.response;
    },
    Swal: {
      isLoading: () => !!state.loading,
      showValidationMessage: (message) => state.validation.push(message),
      fire: async (options) => {
        state.dialogs.push(options);
        if (!options.preConfirm || !state.confirm) return { isConfirmed: false };
        state.loading = true;
        const value = await options.preConfirm();
        state.loading = false;
        return value === false ? { isConfirmed: false } : { isConfirmed: true, value };
      },
    },
  };
  vm.createContext(context);
  for (const [, script] of template.matchAll(/<script>([\s\S]*?)<\/script>/g)) {
    vm.runInContext(script.replaceAll('{{ .IsArchive }}', String(isArchive)), context);
  }
  state.manager = context.partnerManager();
  state.renderActions = (row) => state.config.columns.find((column) => column.data === 'id').render(row.id, 'display', row);
  return state;
}

test('permanent-delete button is visible only on eligible archive rows', () => {
  for (const isArchive of [true, false]) {
    const page = loadPartnerPage(isArchive);
    for (const eligibility of [true, false, undefined, 'true']) {
      const html = page.renderActions({ id: 7, name: 'Partner Uji', can_delete_permanent: eligibility });
      assert.equal(html.includes('data-action="delete-permanent"'), isArchive && eligibility === true);
    }
  }
});

test('delete action passes the selected row as event data and ignores ineligible rows', () => {
  const page = loadPartnerPage();
  const name = 'Partner "Uji" <b>Nama</b>';
  page.row = { id: 7, name, can_delete_permanent: true };
  page.clickDelete.call({});
  assert.equal(page.events.length, 1);
  assert.equal(page.events[0].type, 'req-delete-permanent');
  assert.equal(page.events[0].detail.id, 7);
  assert.equal(page.events[0].detail.name, name);
  page.row.can_delete_permanent = false;
  page.clickDelete.call({});
  assert.equal(page.events.length, 1);
});

test('cancelling confirmation makes no request and releases the pending guard', async () => {
  const page = loadPartnerPage();
  await page.manager.confirmDeletePermanent({ id: 7, name: '<b>Partner Uji</b>' });
  assert.equal(page.requests.length, 0);
  assert.equal(page.manager.isDeletingPermanent, false);
  assert.equal(page.dialogs[0].html, undefined);
  assert.match(page.dialogs[0].text, /<b>Partner Uji<\/b>/);
  assert.equal(page.dialogs[0].focusCancel, true);
  assert.equal(page.dialogs[0].showLoaderOnConfirm, true);
  page.loading = true;
  assert.equal(page.dialogs[0].allowOutsideClick(), false);
  assert.equal(page.dialogs[0].allowEscapeKey(), false);
});

test('confirmed success deletes the exact partner and refreshes the table', async () => {
  const page = loadPartnerPage();
  page.confirm = true;
  page.response = { ok: true, status: 200, json: async () => ({ status: true, message: 'Partner berhasil dihapus permanen' }) };
  await page.manager.confirmDeletePermanent({ id: 7, name: 'Partner Uji' });
  assert.equal(page.requests.length, 1);
  assert.equal(page.requests[0][0], '/master/partner/7/delete-permanent');
  assert.equal(page.requests[0][1].method, 'POST');
  assert.equal(page.reloads.length, 1);
  assert.equal(page.validation.length, 0);
  assert.equal(page.dialogs[1].icon, 'success');
  assert.equal(page.manager.isDeletingPermanent, false);
});

test('server rejection refreshes stale eligibility without displaying success', async () => {
  for (const status of [409, 404, 500]) {
    const page = loadPartnerPage();
    page.confirm = true;
    page.response = { ok: false, status, json: async () => ({ status: false, message: 'Penghapusan ditolak' }) };
    await page.manager.confirmDeletePermanent({ id: 7, name: 'Partner Uji' });
    assert.deepEqual(page.validation, ['Penghapusan ditolak']);
    assert.equal(page.reloads.length, status === 500 ? 0 : 1);
    assert.equal(page.dialogs.length, 1);
    assert.equal(page.manager.isDeletingPermanent, false);
  }
});

test('network failure preserves the table and reports a recoverable error', async () => {
  const page = loadPartnerPage();
  page.confirm = true;
  page.fetchError = new Error('network unavailable');
  await page.manager.confirmDeletePermanent({ id: 7, name: 'Partner Uji' });
  assert.equal(page.reloads.length, 0);
  assert.deepEqual(page.validation, ['Gagal menghubungi server. Silakan coba lagi.']);
  assert.equal(page.manager.isDeletingPermanent, false);
});

test('repeated clicks do not open another dialog while deletion is pending', async () => {
  const page = loadPartnerPage();
  const first = page.manager.confirmDeletePermanent({ id: 7, name: 'Partner Uji' });
  const repeated = page.manager.confirmDeletePermanent({ id: 8, name: 'Partner Lain' });
  await Promise.all([first, repeated]);
  assert.equal(page.dialogs.length, 1);
  assert.equal(page.requests.length, 0);
  assert.equal(page.manager.isDeletingPermanent, false);
});
