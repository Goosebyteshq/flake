const assert = require('assert');

describe('sample', () => {
  it('pass case', () => {
    assert.strictEqual(1, 1);
  });

  it('fail case', () => {
    assert.strictEqual(1, 2);
  });
});
