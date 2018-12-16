var GITHUB_TOKEN = PropertiesService.getScriptProperties().getProperty('github_token');
var TRAVIS_TOKEN = PropertiesService.getScriptProperties().getProperty('travis_token');

var SHEET_ID = PropertiesService.getScriptProperties().getProperty('sheet_id');
var COLS = 16;
var spreadsheet = SHEET_ID ? SpreadsheetApp.openById(SHEET_ID) : SpreadsheetApp.getActiveSpreadsheet();

var SLACK_WEBHOOK_URL = PropertiesService.getScriptProperties().getProperty('slack_webhook_url');

var REPO = 'tarao/prisoners-switch';
var API_REPO = 'https://api.github.com/repos/' + REPO;
var RESULT_TAG = 'verifier.result:';
var BUILD_URL = 'https://travis-ci.com/' + REPO + '/builds/';
var GITHUB_PAGES_URL = 'https://tarao.github.io/prisoners-switch/';

var CHANGED_FILES_VERIFICATION = {
  ok: {
    state: 'success',
    message: 'Changed files seem to be OK'
  },
  violation: {
    state: 'failure',
    message: 'Violation: files outside of `strategy/` have been changed'
  },
  too_many_files: {
    state: 'error',
    message: 'Too many files have been changed'
  },
  error: {
    state: 'error',
    message: 'Verification on changed files failed due to some error'
  }
};

function doPost(e) {
  try {
    if (e.parameter.action === 'report') {
      // https://docs.travis-ci.com/user/notifications#webhooks-delivery-format
      postReport(JSON.parse(e.parameter.payload));
    }

    debug(JSON.stringify({
      parameter: e.parameter,
      query: e.queryString,
      postData: e.postData ? e.postData.getDataAsString() : null
    }));
  } catch (err) {
    debug(JSON.stringify({
      error: err,
      query: e.queryString,
      parameter: e.parameter,
      postData: e.postData ? e.postData.getDataAsString() : null
    }));
  }
}

function doUpdateRanks() {
  try {
    updateRanks();
  } catch (err) {
    debug(JSON.stringify({ error: err }));
  }
}

function postReport(payload) {
  var buildId = payload.id;
  if (!buildId) throw 'No build ID specified';

  // The build information should be retrieved by our own since we
  // cannot trust information from the webhook payload which may be
  // invented by an attacker.
  var build = getBuild(buildId);
  if (!build) throw 'Could not retrieve the build of build ID ' + buildId;
  if (build.repository.slug !== REPO) throw 'Wrong repository';
  if (!build.pull_request_number) throw 'Not a pull request';

  // Retrieve pull request to get head commit SHA1 of the pull request.
  var pullRequest = getPullRequest(build.pull_request_number);
  if (!pullRequest) throw 'No pull request found';
  var sha = pullRequest.head.sha;
  if (!sha) throw 'Commit not found';

  // Verify that the pull request doesn't touch the testing framework
  // or otherwise we cannot trust the test result.
  var tag = verify('/pulls/' + build.pull_request_number);
  if (tag !== 'ok') {
    var result = CHANGED_FILES_VERIFICATION[tag];
    report(result.state, result, pullRequest, buildId);
    return;
  }

  // Retrieve test result from the build log.
  var result = getResult(build.jobs[0].id);
  if (!result) {
    report('error', { message: 'No result found' }, pullRequest, buildId);
    return;
  }

  if (result.success) {
    report('success', result, pullRequest, buildId);
  } else {
    report('failure', result, pullRequest, buildId);
  }
}

function updateRanks() {
  var sh = prepareAnswersSheet();
  var lastRow = sh.getLastRow();
  if (lastRow <= 1) return;

  var lock = lockSheet();
  if (!lock) {
    debug(JSON.stringify({
      error: 'Failed to update ranks due to lock',
      answer: answer
    }));
    return;
  }

  var answers = [];
  var limit = 300;

  try {
    var range = sh.getRange(1+1, COLS, lastRow-1, 2);
    var values = range.getDisplayValues().map(function(row, i) {
      return [ i + 1+1, row[0], row[1] ];
    }).filter(function(item) {
      return item[1] !== item[2];
    });

    if (values.length > 1000) {
      values = shuffle(values).slice(0, limit);
    }

    values.forEach(function(item) {
      var rows = sh.getRange(item[0], 1, 1, COLS-1).getValues();
      var row = rows[0];
      answers.push({
        pullRequestNumber: row[0],
        updatedAt: row[1],
        state: row[2],
        score: row[3],
        steps: row[4],
        usedSwitches: row[5],
        message: row[6],
        user: row[7],
        avatar: row[8],
        url: row[9],
        sha: row[10],
        headRepo: row[11],
        headBranch: row[12],
        baseBranch: row[13],
        build: row[14],
      });
    });
  } finally {
    lock.releaseLock();
  }

  answers.forEach(function(answer) {
    updateAnswer(answer);
  });
}

function report(state, result, pullRequest, buildId) {
  var answer = {
    pullRequestNumber: pullRequest.number,
    updatedAt: pullRequest.updated_at,
    state: state,
    score: result.score || 0,
    steps: result.steps || 0,
    usedSwitches: result.used_switches || 0,
    message: result.message,
    user: pullRequest.user.login,
    avatar: pullRequest.user.avatar_url,
    url: pullRequest.html_url,
    sha: pullRequest.head.sha,
    headRepo: pullRequest.head.repo.full_name,
    headBranch: pullRequest.head.ref,
    baseBranch: pullRequest.base.ref,
    build: BUILD_URL + buildId
  };
  var rank = updateAnswer(answer);
  answer.rank = rank;
  reportToSlack(answer);
}

function updateAnswer(answer) {
  var rank = writeToSheet(answer) || '-';
  var link = GITHUB_PAGES_URL + '?' + [
    [ 'pr', answer.pullRequestNumber ],
    [ 'url', answer.url ],
    [ 'user', answer.user ],
    [ 'avatar', answer.avatar ],
    [ 'state', answer.state ],
    [ 'score', answer.score ],
    [ 'steps', answer.steps ],
    [ 'sw', answer.usedSwitches ],
    [ 'msg', answer.message ],
    [ 'rank', rank ],
    [ 'timestamp', answer.updatedAt ]
  ].map(function(pair) {
    return pair[0] + '=' + encodeURIComponent(pair[1]);
  }).join('&');

  var message = answer.state === 'success' ?
      'Score: ' + answer.score :
      answer.message;
  updateResultStatus(answer.sha, answer.state, message, link);
  return rank;
}

function getBuild(buildId) {
  var res = travisRequest('/build/' + buildId, {
    headers: { 'Accept': 'text/plain' }
  });
  var code = res.getResponseCode();
  if (200 <= code && code < 300) {
    return JSON.parse(res.getContentText());
  }
}

function getResult(jobId) {
  for (var i=0; i < 10; i++) {
    try {
      var res = travisRequest('/job/' + jobId + '/log', {
        headers: { 'Accept': 'text/plain' }
      });
      var resultLine = res.getContentText().split(/[\r\n]+/).filter(function(line) {
        return line.indexOf(RESULT_TAG) == 0;
      })[0];

      if (!resultLine) throw 'no result';
      return JSON.parse(resultLine.substring(RESULT_TAG.length));

    } catch (e) {
      // Wait for Travis CI to output the log.
      Utilities.sleep(5000);
      continue;
    }
  }
}

function getPullRequest(pullRequestNumber) {
  var res = githubRequest('/pulls/' + pullRequestNumber);
  return JSON.parse(res.getContentText());
}

function verify(pullRequestPath) {
  try {
    var res = githubRequest(pullRequestPath + '/files');
    var files = JSON.parse(res.getContentText());
    if (files.length > 20) return 'too_many_files';
    return files.every(function(file) {
      return file.filename.indexOf('strategy/') == 0;
    }) ? 'ok' : 'violation';
  } catch (e) {
    return 'error';
  }
}

function updateResultStatus(sha, state, description, link) {
  githubRequest('/statuses/' + sha, {
    method: 'POST',
    payload: JSON.stringify({
      state: state,
      description: description,
      target_url: link,
      context: 'Result'
    })
  });
}

function travisRequest(path, options) {
  var url = 'https://api.travis-ci.com' + path;
  options = options || {};
  options.headers = options.headers || {};
  options.headers['Authorization'] = 'token ' + TRAVIS_TOKEN;
  options.headers['Travis-API-Version'] = '3';
  return UrlFetchApp.fetch(url, options);
}

function githubRequest(path, options) {
  var url = API_REPO + path;
  options = options || {};
  options.headers = options.headers || {};
  options.headers['Authorization'] = 'token ' + GITHUB_TOKEN;
  options.contentType = 'application/json';
  return UrlFetchApp.fetch(url, options);
}

function prepareAnswersSheet() {
  var sheetName = 'answers';
  var sh = spreadsheet.getSheetByName(sheetName);
  if (sh == null) {
    var activeSh = spreadsheet.getActiveSheet();
    sheetNum = spreadsheet.getSheets().length;
    sh = spreadsheet.insertSheet(sheetName, sheetNum);
    sh.getRange(1, 1, 1, COLS+1).setValues([[
      'PR#', 'Timestamp', 'State', 'Score', 'Steps', 'Sw', 'Message',
      'Author', 'Avatar',
      'Pull Request', 'SHA', 'Repository', 'Branch', 'Base', 'Build',
      'Rank', 'Previous Rank'
    ]]).setBackground('#cfe2f3').setFontWeight('bold');
  }
  spreadsheet.setActiveSheet(sh);
  return sh;
}

function lockSheet() {
  var docLock = LockService.getDocumentLock();
  var i = 0;
  for (var i=0; i < 20; i++) {
    if (docLock.tryLock(500)) return docLock;
  }
}

function writeToSheet(answer) {
  var lock = lockSheet();
  if (!lock) {
    debug(JSON.stringify({
      error: 'Failed to report due to lock',
      answer: answer
    }));
    return;
  }

  try {
    var sh = prepareAnswersSheet();
    var lastRow = sh.getLastRow();
    var targetRange = (function() {
      if (lastRow > 1) {
        var values = sh.getRange(1+1, 1, lastRow-1, 1).getValues();
        var found = values.map(function(cols, row) {
          return [ row, cols[0] ];
        }).filter(function(item) {
          return item[1] == answer.pullRequestNumber;
        })[0];
        if (found) return sh.getRange(1+1+found[0], 1, 1, COLS);
      }
      return sh.insertRowAfter(lastRow).getRange(lastRow+1, 1, 1, COLS);
    })();
    var row = targetRange.getRow();
    targetRange.setValues([[
      answer.pullRequestNumber,
      answer.updatedAt,
      answer.state,
      answer.score,
      answer.steps,
      answer.usedSwitches,
      answer.message,
      answer.user,
      answer.avatar,
      answer.url,
      answer.sha,
      answer.headRepo,
      answer.headBranch,
      answer.baseBranch,
      answer.build,
      '=IF($C' + row + '="success",RANK($D' + row + ',$D$2:$D,0),IF($C' + row + '="","","-"))'
    ]]).clearFormat();

    var rank = targetRange.offset(0, COLS-1, 1, 1).getDisplayValue()

    // Write the current value of the calculated rank so that we will
    // know if the rank has been changed in future.
    targetRange.offset(0, COLS, 1, 1).setValues([[rank]]).clearFormat();

    return rank;

  } finally {
    lock.releaseLock();
  }
}

function reportToSlack(answer) {
  UrlFetchApp.fetch(SLACK_WEBHOOK_URL, {
    method: 'POST',
    payload: JSON.stringify({
      username: 'The Prisoners and the Switch Room Problem',
      icon_emoji: ':bulb:',
      attachments: [ {
        pretext: 'A new answer posted',
        color: answer.state === 'success' ? '#39aa56' : '#db4545',
        author_name: answer.user,
        author_link: 'https://github.com/' + answer.user + '/',
        author_icon: answer.avatar,
        title: '#' + answer.pullRequestNumber,
        title_link: answer.url,
        text: answer.message,
        fields: [
          {
            title: 'Score',
            value: answer.score,
            short: true
          },
          {
            title: 'Rank',
            value: answer.rank || '-',
            short: true
          },
          {
            title: 'Steps',
            value: answer.steps,
            short: true
          },
          {
            title: 'Used Switches',
            value: answer.usedSwitches,
            short: true
          }
        ],
        footer: REPO,
        footer_icon: "https://github.com/fluidicon.png"
      } ]
    })
  });
}

//util

function shuffle(array) {
  for (var i = array.length - 1; i > 0; i--) {
    var j = Math.floor(Math.random() * (i + 1));
    var tmp = array[i];
    array[i] = array[j];
    array[j] = tmp;
  }
  return array;
}

//logger

function prepareLogSheet() {
  var sheetName = 'log';
  var sh = spreadsheet.getSheetByName(sheetName);
  if (sh == null) {
    var activeSh = spreadsheet.getActiveSheet();
    sheetNum = spreadsheet.getSheets().length;
    sh = spreadsheet.insertSheet(sheetName, sheetNum);
    sh.getRange('A1:C1').setValues([[
      'timestamp', 'level', 'message'
    ]]).setBackground('#cfe2f3').setFontWeight('bold');
    sh.getRange('A2:C2').setValues([[
      new Date().toISOString(),
      'info',
      sheetName + ' has been created.'
    ]]).clearFormat();

    spreadsheet.setActiveSheet(activeSh);
  }
  return sh;
}

function putLog(level, message) {
  var sh = prepareLogSheet();
  var now = new Date();
  var lastRow = sh.getLastRow();
  sh.insertRowAfter(lastRow).getRange(lastRow+1, 1, 1, 3).setValues([[
    now.toISOString(),
    level,
    message
  ]]);
  return sh;
}

function debug(message) {
  putLog('debug', message);
}
