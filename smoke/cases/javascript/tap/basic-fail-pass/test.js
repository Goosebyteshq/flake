const test = require('tape');

test('pass case', (t) => {
  t.equal(1, 1, 'pass case');
  t.end();
});

test('fail case', (t) => {
  t.equal(1, 2, 'fail case');
  t.end();
});
