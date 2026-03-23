<script>
  export let customerName = "Sample Customer";
  export let amount = 50000;
  export let primaryColor = "#2563eb";

  // New Modular State with Drag and Drop
  let blocks = [
    { id: 'header', name: 'Header (Logo & Brand)', enabled: true },
    { id: 'dispatch', name: 'Logistics & Dispatch Info', enabled: true },
    { id: 'billto', name: 'Bill To Details', enabled: true },
    { id: 'custom', name: 'Dynamic Fields', enabled: true },
    { id: 'items', name: 'Items Table', enabled: true },
    { id: 'hsn_summary', name: 'HSN/SAC Tax Summary', enabled: true },
    { id: 'totals', name: 'Totals & Tax', enabled: true },
    { id: 'signature', name: 'Signature Pad', enabled: true },
    { id: 'qrcode', name: 'UPI QR Code', enabled: true }
  ];

  let logistics = {
    transporter: "FlyFleet Logistics",
    vehicle_no: "DL-01-LH-8783",
    lr_no: "17261920",
    eway_bill: "751590248946",
    place_of_supply: "Uttar Pradesh (09)",
    reverse_charge: "No"
  };

  let customFields = [
    { key: 'Driver Name', value: 'Rajesh Kumar' }
  ];

  let selectedLogo = 'logo1.jpg';
  let logos = [
    { id: 'none', label: 'No Logo' },
    { id: 'logo1.jpg', label: 'Primary Logo' },
    { id: 'logo2.jpg', label: 'Secondary Logo' }
  ];

  let documentType = 'tax_invoice';
  let documentTypes = [
    { id: 'tax_invoice', label: 'Tax Invoice' },
    { id: 'delivery_challan', label: 'Delivery Challan (Job Work)' }
  ];

  $: gst = documentType === 'tax_invoice' ? amount * 0.18 : 0;
  $: total = amount + gst;

  // Drag and Drop State
  let draggingIndex = null;

  function toggleBlock(id) {
    blocks = blocks.map(b => b.id === id ? { ...b, enabled: !b.enabled } : b);
  }

  function addField() {
    customFields = [...customFields, { key: '', value: '' }];
  }

  function removeField(index) {
    customFields = customFields.filter((_, i) => i !== index);
  }

  function handleDragStart(event, index) {
    draggingIndex = index;
    event.dataTransfer.effectAllowed = 'move';
    event.dataTransfer.setData('text/plain', index);
  }

  function handleDragOver(event, index) {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'move';
  }

  function handleDrop(event, targetIndex) {
    event.preventDefault();
    if (draggingIndex !== null && draggingIndex !== targetIndex) {
      let newBlocks = [...blocks];
      const itemToMove = newBlocks[draggingIndex];
      newBlocks.splice(draggingIndex, 1);
      newBlocks.splice(targetIndex, 0, itemToMove);
      blocks = newBlocks;
    }
    draggingIndex = null;
  }

  async function handleSave() {
    console.log("Saving enterprise template configuration...");
    const config = {
      primary_color: primaryColor,
      document_type: documentType,
      selected_logo: selectedLogo,
      enabled_blocks: blocks.filter(b => b.enabled).map(b => b.id),
      block_order: blocks.map(b => b.id),
      custom_fields: customFields
    };

    try {
      const response = await fetch("/api/v1/billing/settings", {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          render_config: JSON.stringify(config)
        })
      });

      if (response.ok) {
        alert("✅ Enterprise Layout Persisted to System!");
      } else {
        const err = await response.json();
        alert("❌ Failed to save: " + (err.error || "Unknown error"));
      }
    } catch (e) {
      console.error("Save error:", e);
      alert("❌ Network Error: Could not connect to backend.");
    }
  }
</script>

<div class="builder-container" style="--brand-color: {primaryColor}">
  <div class="sidebar">
    <h3>Invoice Layout</h3>
    <p class="subtitle">Drag to Reorder • Click to Toggle</p>
    
    <div class="tools-grid">
      {#each blocks as block, index}
        <div 
          class="tool-item {block.enabled ? 'active' : ''}" 
          draggable="true"
          on:dragstart={(e) => handleDragStart(e, index)}
          on:dragover={(e) => handleDragOver(e, index)}
          on:drop={(e) => handleDrop(e, index)}
        >
          <span class="drag-handle">☰</span>
          <button class="toggle-btn" on:click={() => toggleBlock(block.id)}>
            <span class="status-dot"></span>
            {block.name}
          </button>
        </div>
      {/each}
    </div>
    
    <hr />
    
    <div class="controls">
      <label>Document Type:</label>
      <select bind:value={documentType}>
        {#each documentTypes as type}
          <option value={type.id}>{type.label}</option>
        {/each}
      </select>

      <label>Brand Logo:</label>
      <select bind:value={selectedLogo}>
        {#each logos as logo}
          <option value={logo.id}>{logo.label}</option>
        {/each}
      </select>

      <label>Preview Amount (₹):</label>
      <input type="number" bind:value={amount} />
      
      <label>Brand Color:</label>
      <input type="color" bind:value={primaryColor} />

      <hr style="margin: 20px 0" />
      <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 5px;">
        <label style="margin:0">Custom Data Fields:</label>
        <button class="small-btn" style="background:#3b82f6" on:click={addField}>+ Add Field</button>
      </div>
      {#each customFields as field, i}
        <div style="display: flex; gap: 5px; margin-bottom: 5px;">
          <input style="margin:0" placeholder="Key (e.g. Route)" bind:value={field.key} />
          <input style="margin:0" placeholder="Value" bind:value={field.value} />
          <button class="small-btn" style="background:#ef4444" on:click={() => removeField(i)}>✕</button>
        </div>
      {/each}
    </div>

    <button on:click={handleSave} class="save-btn">Apply to System</button>
  </div>

  <div class="preview-area">
    <div class="invoice-paper">
      {#each blocks as block}
        {#if block.enabled}
          {#if block.id === 'header'}
            <header style="background: {primaryColor}">
              <div class="header-content">
                {#if selectedLogo !== 'none'}
                  <!-- Note: In Termux we simulate the logo preview -->
                  <div class="logo-placeholder">[{selectedLogo}]</div>
                {/if}
                <div>
                  <h1>FLYFLEET</h1>
                  <p>{documentType === 'tax_invoice' ? 'Official Tax Invoice' : 'Delivery Challan (Job Work)'}</p>
                </div>
              </div>
            </header>
          {/if}

          {#if block.id === 'dispatch'}
            <div class="content-padding pt-0">
              <div class="dispatch-grid">
                <div class="dispatch-item"><span>Transporter:</span> {logistics.transporter}</div>
                <div class="dispatch-item"><span>Vehicle No:</span> {logistics.vehicle_no}</div>
                <div class="dispatch-item"><span>LR/RR No:</span> {logistics.lr_no}</div>
                <div class="dispatch-item"><span>E-Way Bill:</span> {logistics.eway_bill}</div>
                <div class="dispatch-item"><span>Place of Supply:</span> {logistics.place_of_supply}</div>
                <div class="dispatch-item"><span>Reverse Charge:</span> {logistics.reverse_charge}</div>
              </div>
            </div>
          {/if}

          {#if block.id === 'billto'}
            <section class="address-section content-padding pt-0">
              <div class="address-box">
                <span class="label">CONSIGNEE (SHIP TO)</span>
                <div class="address-content">
                  <strong>{customerName}</strong><br/>
                  H-45, Sector-63, Noida, UP<br/>
                  GSTIN: 09AAAAA0000A1Z5
                </div>
              </div>
              <div class="address-box">
                <span class="label">BUYER (BILL TO)</span>
                <div class="address-content">
                  <strong>{customerName}</strong><br/>
                  Gautambuddha Nagar, UP<br/>
                  GSTIN: 09AAAAA0000A1Z5
                </div>
              </div>
            </section>
          {/if}

          {#if block.id === 'items'}
            <div class="content-padding pt-0">
              <table class="item-table industrial">
                <thead>
                  <tr>
                    <th>Description of Goods</th>
                    <th>HSN/SAC</th>
                    <th class="text-right">Qty</th>
                    <th class="text-right">Rate</th>
                    <th class="text-right">Amount</th>
                  </tr>
                </thead>
                <tbody>
                  <tr>
                    <td>Logistics Freight Service (Industrial)</td>
                    <td>996511</td>
                    <td class="text-right">1</td>
                    <td class="text-right">₹{amount.toLocaleString()}</td>
                    <td class="text-right">₹{amount.toLocaleString()}</td>
                  </tr>
                </tbody>
              </table>
            </div>
          {/if}

          {#if block.id === 'hsn_summary'}
            <div class="content-padding pt-0">
              <div class="hsn-summary">
                <span class="label">HSN/SAC TAX SUMMARY</span>
                <table class="mini-table">
                  <thead>
                    <tr>
                      <th>HSN</th>
                      <th>Taxable</th>
                      <th>CGST</th>
                      <th>SGST</th>
                      <th>IGST</th>
                      <th class="text-right">Total</th>
                    </tr>
                  </thead>
                  <tbody>
                    <tr>
                      <td>996511</td>
                      <td>₹{amount.toLocaleString()}</td>
                      <td>₹{(gst/2).toLocaleString()}</td>
                      <td>₹{(gst/2).toLocaleString()}</td>
                      <td>₹0</td>
                      <td class="text-right">₹{gst.toLocaleString()}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          {/if}

          {#if block.id === 'qrcode'}
            <div class="content-padding pb-0">
              <div class="qr-block">
                <div class="qr-mock">QR</div>
                <span>Scan to Pay UPI</span>
              </div>
            </div>
          {/if}

          {#if block.id === 'signature'}
            <div class="content-padding pb-0">
              <div class="sig-block">
                <div class="sig-line"></div>
                <span>Authorized Signatory</span>
              </div>
            </div>
          {/if}

          {#if block.id === 'totals'}
            <div class="content-padding pt-0">
              <div class="totals">
                {#if documentType === 'tax_invoice'}
                  <div class="row"><span>Subtotal</span><span>₹{amount.toLocaleString()}</span></div>
                  <div class="row"><span>GST (18%)</span><span>₹{gst.toLocaleString()}</span></div>
                  <div class="row grand" style="color: {primaryColor}">
                    <span>Total</span><span>₹{total.toLocaleString()}</span>
                  </div>
                {:else}
                  <div class="row grand" style="color: {primaryColor}">
                    <span>Total Value</span><span>₹{amount.toLocaleString()}</span>
                  </div>
                {/if}
              </div>
            </div>
          {/if}
        {/if}
      {/each}
    </div>
  </div>
</div>

<style>
  /* ... existing styles ... */
  .builder-container {
    display: grid;
    grid-template-columns: 320px 1fr;
    height: 750px;
    background: #f8fafc;
    color: #1e293b;
    border-radius: 12px;
    overflow: hidden;
  }
  .sidebar { background: #fff; padding: 24px; border-right: 1px solid #e2e8f0; display: flex; flex-direction: column; overflow-y: auto; }
  .subtitle { font-size: 0.8em; color: #64748b; margin-top: -10px; margin-bottom: 20px; }
  
  .tools-grid { display: grid; gap: 8px; }
  .tool-item {
    display: flex;
    align-items: center;
    background: #f1f5f9;
    border: 1px solid #e2e8f0;
    border-radius: 8px;
    transition: all 0.2s;
  }
  .tool-item.active { background: #f0fdf4; border-color: #22c55e; color: #166534; }
  .drag-handle { padding: 12px; cursor: grab; color: #94a3b8; border-right: 1px solid #e2e8f0; }
  .tool-item.active .drag-handle { border-color: #bbf7d0; color: #22c55e; }
  .toggle-btn {
    flex-grow: 1; padding: 12px; border: none; background: transparent; cursor: pointer;
    display: flex; align-items: center; gap: 10px; font-weight: 600; font-size: 0.85em; text-align: left; color: inherit;
  }
  
  .status-dot { width: 8px; height: 8px; border-radius: 50%; background: #cbd5e1; }
  .active .status-dot { background: #22c55e; }

  .preview-area { padding: 40px; display: flex; justify-content: center; background: #e2e8f0; overflow-y: auto; }
  .invoice-paper { width: 100%; max-width: 600px; background: white; box-shadow: 0 20px 25px -5px rgba(0,0,0,0.1); min-height: 800px; display: flex; flex-direction: column; }
  
  header { padding: 30px; color: white; text-align: left; }
  .header-content { display: flex; justify-content: space-between; align-items: center; }
  .logo-placeholder { background: rgba(255,255,255,0.2); padding: 20px; border-radius: 8px; font-size: 0.8em; font-weight: bold; }
  
  .content-padding { padding: 20px 30px; }
  .pt-0 { padding-top: 0; }
  .pb-0 { padding-bottom: 0; }
  
  .label { font-size: 0.7em; font-weight: 800; color: #94a3b8; }
  .item-table { width: 100%; border-collapse: collapse; margin: 10px 0; }
  .item-table th { text-align: left; padding: 10px; background: #f8fafc; border-bottom: 2px solid #e2e8f0; font-size: 0.8em; }
  .item-table td { padding: 15px 10px; border-bottom: 1px solid #f1f5f9; }
  
  .qr-block { display: flex; align-items: center; gap: 10px; }
  .qr-mock { width: 60px; height: 60px; border: 2px solid #e2e8f0; display: flex; align-items: center; justify-content: center; font-weight: bold; color: #cbd5e1; }
  .qr-block span { font-size: 0.7em; color: #64748b; }

  .sig-block { margin-top: 10px; display: flex; flex-direction: column; align-items: flex-end; }
  .sig-line { width: 150px; border-bottom: 2px solid #1e293b; margin-bottom: 5px; }
  .sig-block span { font-size: 0.7em; font-weight: bold; }

  .dispatch-grid { display: grid; grid-template-columns: 1fr 1fr; border: 1px solid #e2e8f0; border-radius: 6px; padding: 10px; gap: 5px; background: #f8fafc; }
  .dispatch-item { font-size: 0.75em; color: #1e293b; }
  .dispatch-item span { font-weight: bold; color: #64748b; margin-right: 5px; }

  .address-section { display: flex; gap: 20px; }
  .address-box { flex: 1; border: 1px solid #e2e8f0; padding: 15px; border-radius: 6px; }
  .address-content { font-size: 0.8em; line-height: 1.4; color: #1e293b; margin-top: 5px; }

  .item-table.industrial th { font-size: 0.7em; text-transform: uppercase; letter-spacing: 0.05em; }
  .item-table.industrial td { font-size: 0.85em; }

  .hsn-summary { border: 1px solid #e2e8f0; padding: 15px; border-radius: 6px; margin-top: 10px; }
  .mini-table { width: 100%; border-collapse: collapse; margin-top: 10px; }
  .mini-table th { font-size: 0.65em; text-align: left; padding: 5px; border-bottom: 1px solid #e2e8f0; color: #64748b; }
  .mini-table td { font-size: 0.75em; padding: 8px 5px; border-bottom: 1px dotted #f1f5f9; }

  .custom-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; background: #f8fafc; padding: 15px; border-radius: 8px; border: 1px solid #e2e8f0; }
  .custom-item { font-size: 0.85em; }
  .custom-key { font-weight: bold; color: #64748b; }
  .custom-val { color: #1e293b; margin-left: 5px; }
  
  .small-btn { border: none; border-radius: 4px; cursor: pointer; padding: 6px 10px; color: white; font-weight: bold; font-size: 0.75em;}

  .totals { width: 300px; margin-left: auto; }
  .totals .row { display: flex; justify-content: space-between; padding: 8px 0; font-size: 0.9em; }
  .totals .grand { border-top: 2px solid #e2e8f0; margin-top: 10px; padding-top: 10px; font-weight: 800; font-size: 1.1em; }

  .save-btn { margin-top: 20px; padding: 16px; background: #1e293b; color: white; border: none; border-radius: 10px; cursor: pointer; font-weight: bold; box-shadow: 0 4px 6px -1px rgba(0,0,0,0.1); }
  
  .controls { margin-top: 10px; }
  label { display: block; font-size: 0.8em; font-weight: bold; margin-bottom: 5px; color: #475569; }
  input, select { width: 100%; margin-bottom: 15px; padding: 10px; border: 1px solid #e2e8f0; border-radius: 6px; box-sizing: border-box; }
  .text-right { text-align: right; }
</style>
