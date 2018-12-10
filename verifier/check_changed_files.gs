var TOKEN = PropertiesService.getScriptProperties().getProperty('token');
var GITHUB_TOKEN = PropertiesService.getScriptProperties().getProperty('github_token');
var SHEET_ID = PropertiesService.getScriptProperties().getProperty('sheet_id');
var spreadsheet = SHEET_ID ? SpreadsheetApp.openById(SHEET_ID) : SpreadsheetApp.getActiveSpreadsheet();

function doPost(e) {
  if (e.parameter.token !== TOKEN) return;

  var json = e.postData.getDataAsString();
  var payload = JSON.parse(json);

  if (e.parameter.action === 'verify') {
    verify(payload);
  }
}

function githubRequest(url, options) {
  options = options || {};
  options.headers = options.headers || {};
  options.headers['Authorization'] = 'token ' + GITHUB_TOKEN;
  options.contentType = 'application/json';
  return UrlFetchApp.fetch(url, options);
}

function verify(payload) {
  var files_url = payload.pull_request.url + '/files';
  var statuses_url = payload.pull_request.statuses_url;
  var results = {
    ok: {
      state: 'success',
      description: 'Changed files seem to be OK'
    },
    violation: {
      state: 'failure',
      description: 'Violation: files outside of `strategy/` have been changed'
    },
    too_many_files: {
      state: 'error',
      description: 'Too many files have been changed'
    },
    error: {
      state: 'error',
      description: 'Verification on changed files failed due to some error'
    }
  };

  var tag = (function() {
    try {
      var res = githubRequest(files_url);
      var files = JSON.parse(res.getContentText());
      if (files.length > 20) return 'too_many_files';
      return files.every(function(file) {
        return file.filename.startsWith('strategy/');
      }) ? 'ok' : 'violation';
    } catch (e) {
      return 'error';
    }
  })();
  var result = results[tag];

  githubRequest(statuses_url, {
    method: 'POST',
    payload: JSON.stringify({
      state: result.state,
      description: result.description,
      target_url: 'https://github.com/tarao/prisoners-switch/blob/master/verifier/check_changed_files.gs',
      context: 'changed-files'
    })
  });
}

//logger

function log_sheet_() {
  var sheet_name = 'log';
  var sh = spreadsheet.getSheetByName(sheet_name);
  if (sh == null) {
    var active_sh = spreadsheet.getActiveSheet();
    sheet_num = spreadsheet.getSheets().length;
    sh = spreadsheet.insertSheet(sheet_name, sheet_num);
    sh.getRange('A1:C1').setValues([['timestamp', 'level', 'message']]).setBackground('#cfe2f3').setFontWeight('bold');
    sh.getRange('A2:C2').setValues([[new Date(), 'info', sheet_name + ' has been created.']]).clearFormat();

    spreadsheet.setActiveSheet(active_sh);
  }
  return sh;
}

function log_(level, message) {
  var sh = log_sheet_();
  var now = new Date();
  var last_row = sh.getLastRow();
  sh.insertRowAfter(last_row).getRange(last_row+1, 1, 1, 3).setValues([[now, level, message]]);
  return sh;
}

function debug(message) {
  log_('debug', message);
}
