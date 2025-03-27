const { createHmac } = require('crypto');

const secret = 'Yi4GUR8wJHyDpPL0vbhWHwn5pMJuUvL1J35KPvA07FZv8dqwkY';

const streamId = '97424922-2fef-49c1-8272-c8b8ffadc354';

console.log(createHmac('sha256', secret).update(streamId).digest('hex'))
