const axios = require('axios');
const c = axios.create({ baseURL: 'http://localhost:8084/api/v1' });
console.log(c.getUri({url: '/ui/misc/legal_docs'}));
