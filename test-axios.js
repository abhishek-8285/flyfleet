const axios = require('axios');
const c1 = axios.create({ baseURL: 'http://localhost:8084/api/v1' });
console.log("With host + /:", c1.getUri({ url: '/ui/misc/legal_docs' }));
console.log("With host + no /:", c1.getUri({ url: 'ui/misc/legal_docs' }));

const c2 = axios.create({ baseURL: '/api/v1' });
console.log("With path + /:", c2.getUri({ url: '/ui/misc/legal_docs' }));
console.log("With path + no /:", c2.getUri({ url: 'ui/misc/legal_docs' }));
