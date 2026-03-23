const fs = require('fs');
const svelte = require('svelte/compiler');

try {
  const source = fs.readFileSync('apps/dashboard-web/src/components/builder/InvoiceBuilder.svelte', 'utf8');
  const result = svelte.compile(source, {
    generate: 'dom',
    name: 'InvoiceBuilder'
  });

  fs.writeFileSync('apps/dashboard-web/src/components/builder/InvoiceBuilderCompiled.js', '/* eslint-disable */\n// @ts-nocheck\n' + result.js.code);
  if (result.css && result.css.code) {
    fs.writeFileSync('apps/dashboard-web/src/components/builder/InvoiceBuilderCompiled.css', result.css.code);
  }
  console.log("✅ Svelte component compiled successfully with CSS.");
} catch (e) {
  console.error("❌ Failed to compile Svelte component:", e);
  process.exit(1);
}
