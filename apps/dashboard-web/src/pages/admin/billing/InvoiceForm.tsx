import {
    Calendar,
    ChevronLeft,
    IndianRupee,
    Package,
    MapPin,
    Plus,
    Save,
    Trash2,
    User,
    Phone,
    Mail,
} from "lucide-react";
import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useUIStrings } from "../../../hooks/useUIStrings";
import billingService from "../../../services/billingService";
import { invoicingV2Service } from "../../../services/invoicingV2Service";
import CustomerSearch from "../../../components/common/CustomerSearch";
import { Customer } from "../../../services/customerV2Service";
import { Invoice, InvoiceItem, InvoiceStatus } from "../../../types/billing";

const InvoiceForm: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const isEdit = !!id;

  const { uiStrings } = useUIStrings("admin_billing");

  const [formData, setFormData] = useState<Partial<Invoice>>({
    customer_name: "",
    customer_email: "",
    customer_phone: "",
    billing_address: "",
    invoice_date: new Date().toISOString().split("T")[0],
    due_date: new Date(Date.now() + 14 * 24 * 60 * 60 * 1000)
      .toISOString()
      .split("T")[0],
    template_type: "TAX_INVOICE",
    currency: "INR",
    category: "Transportation",
    // Transporter Profile (Master)
    bank_name: "",
    account_no: "",
    ifsc_code: "",
    branch_name: "",
    beneficiary_name: "",
    gstin_seller: "",
    pan_seller: "",
    // Logistics
    vehicle_no: "",
    driver_name: "",
    driver_mobile: "",
    pickup_location: "",
    drop_location: "",
    eway_bill_no: "",
    billed_distance: 0,
    place_of_supply: "KA", // Default to Karnataka
    items: [
      {
        id: "",
        invoice_id: "",
        description: "",
        hsn_sac: "",
        unit: "Trip",
        quantity: 1,
        unit_price: 0,
        tax_percent: 18,
        total: 0,
      },
    ],
  });

  const [loading, setLoading] = useState(false);
  const [billingMeta, setBillingMeta] = useState<any>(null);
  const [showErrors, setShowErrors] = useState(false);

  const isFieldMissing = (value: any) => {
    if (value === null || value === undefined) return true;
    if (typeof value === "string") return value.trim() === "";
    if (Array.isArray(value)) return value.length === 0;
    return false;
  };

  useEffect(() => {
    fetch("/api/v1/ui/global_config")
      .then((res) => res.json())
      .then((data) => {
        setBillingMeta(data.billing_metadata);
        if (!isEdit && data.billing_metadata) {
          setFormData((prev) => ({
            ...prev,
            currency: "INR",
            items: [
              {
                id: "",
                invoice_id: "",
                description: "",
                quantity: 1,
                unit_price: 0,
                tax_percent: data.billing_metadata.default_tax_percent || 18,
                total: 0,
              },
            ],
          }));
        }
      })
      .catch((err) => console.error(err));

    if (isEdit) {
      fetchInvoice();
    }
  }, [id]);

  const fetchInvoice = async () => {
    try {
      const data = await billingService.getInvoice(id!);
      setFormData(data);
    } catch (error) {
      console.error("Failed to fetch invoice", error);
    }
  };

  const handleAddItem = () => {
    setFormData({
      ...formData,
      items: [
        ...(formData.items || []),
        {
          id: "",
          invoice_id: "",
          description: "",
          hsn_sac: "",
          unit: "Trip",
          quantity: 1,
          unit_price: 0,
          tax_percent: billingMeta?.default_tax_percent || 18,
          total: 0,
        },
      ],
    });
  };

  const handleRemoveItem = (index: number) => {
    const newItems = (formData.items || []).filter((_, i) => i !== index);
    setFormData({ ...formData, items: newItems });
  };

  const handleItemChange = (
    index: number,
    field: keyof InvoiceItem,
    value: any,
  ) => {
    const newItems = [...(formData.items || [])];
    newItems[index] = { ...newItems[index], [field]: value };

    // Recalculate item total
    const item = newItems[index];
    const subtotal = item.quantity * item.unit_price;
    item.total = subtotal + subtotal * (item.tax_percent / 100);

    setFormData({ ...formData, items: newItems });
  };

  const calculateSubtotal = () => {
    return (formData.items || []).reduce(
      (acc, item) => acc + item.quantity * item.unit_price,
      0,
    );
  };

  const calculateTax = () => {
    return (formData.items || []).reduce(
      (acc, item) =>
        acc + item.quantity * item.unit_price * (item.tax_percent / 100),
      0,
    );
  };

  const calculateTotal = () => {
    return calculateSubtotal() + calculateTax();
  };

  const handleCustomerSelect = (customer: Customer) => {
    let phone = customer.phone || "";
    // Auto-prefix if it's a 10 digit number and doesn't have a prefix
    if (phone.length === 10 && !phone.startsWith("+")) {
      phone = `+91 ${phone.substring(0, 5)} ${phone.substring(5)}`;
    }

    setFormData((prev) => ({
      ...prev,
      customer_name: customer.name,
      customer_email: customer.email || "",
      customer_phone: phone,
      billing_address: customer.address || "",
    }));
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    // Validate required fields
    const hasMissingFields = 
      isFieldMissing(formData.customer_name) ||
      isFieldMissing(formData.customer_phone) ||
      isFieldMissing(formData.billing_address) ||
      isFieldMissing(formData.invoice_date) ||
      isFieldMissing(formData.due_date) ||
      (formData.items || []).some(item => 
        isFieldMissing(item.description) || 
        item.quantity <= 0 || 
        item.unit_price <= 0
      );

    if (hasMissingFields) {
      setShowErrors(true);
      return;
    }

    setLoading(true);
    try {
      const subtotal = calculateSubtotal();
      const tax = calculateTax();
      const total = calculateTotal();

      const payload = {
        customer_name: formData.customer_name!,
        customer_email: formData.customer_email || "",
        customer_phone: formData.customer_phone!,
        billing_address: formData.billing_address || "",
        invoice_date: formData.invoice_date,
        due_date: formData.due_date,
        currency: formData.currency || "INR",
        template_type: formData.template_type,
        // Logistics
        trip_id: formData.trip_id,
        vehicle_no: formData.vehicle_no,
        driver_name: formData.driver_name,
        driver_mobile: formData.driver_mobile,
        pickup_location: formData.pickup_location,
        drop_location: formData.drop_location,
        pickup_time: formData.pickup_time,
        drop_time: formData.drop_time,
        billed_distance: formData.billed_distance,
        // Compliance
        place_of_supply: formData.place_of_supply,
        reverse_charge: formData.reverse_charge,
        gstin_seller: formData.gstin_seller,
        pan_seller: formData.pan_seller,
        eway_bill_no: formData.eway_bill_no,
        irn: formData.irn,
        // Bank
        bank_name: formData.bank_name,
        account_no: formData.account_no,
        ifsc_code: formData.ifsc_code,
        branch_name: formData.branch_name,
        beneficiary_name: formData.beneficiary_name,

        items: (formData.items || []).map((item) => ({
          description: item.description,
          quantity: item.quantity,
          unit_price: item.unit_price,
          tax_percent: item.tax_percent,
          hsn_sac: item.hsn_sac,
          unit: item.unit,
          list_price: item.unit_price, // Assuming unit price as list price for now
          discount_amount: 0,
        })),
        notes: formData.notes || "",
      };

      if (isEdit) {
        await billingService.updateInvoice(id!, payload as any);
      } else {
        await invoicingV2Service.createInvoice(payload as any);
      }
      navigate("/admin/billing");
    } catch (error: any) {
      console.error("Failed to save invoice", error);
      alert(error?.message || "Failed to save invoice");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="max-w-4xl mx-auto space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <button
          onClick={() => navigate(-1)}
          className="flex items-center gap-2 text-gray-400 hover:text-white transition-colors"
        >
          <ChevronLeft className="w-5 h-5" />
          Back
        </button>
        <h1 className="text-2xl font-bold text-white">
          {isEdit ? "Edit Invoice" : "Create New Invoice"}
        </h1>
        <div className="flex bg-gray-900 border border-white/10 p-1 rounded-xl">
          <button
            type="button"
            onClick={() => setFormData(prev => ({ ...prev, template_type: "TAX_INVOICE" }))}
            className={`px-4 py-1.5 rounded-lg text-xs font-bold transition-all ${formData.template_type === "TAX_INVOICE" ? "bg-primary text-white shadow-lg shadow-primary/20" : "text-gray-500 hover:text-gray-300"}`}
          >
            TAX INVOICE
          </button>
          <button
            type="button"
            onClick={() => setFormData(prev => ({ ...prev, template_type: "TRIP_NOTE" }))}
            className={`px-4 py-1.5 rounded-lg text-xs font-bold transition-all ${formData.template_type === "TRIP_NOTE" ? "bg-primary text-white shadow-lg shadow-primary/20" : "text-gray-500 hover:text-gray-300"}`}
          >
            TRIP NOTE
          </button>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Customer & Dates Section */}
        <div className="bg-gray-900/50 backdrop-blur-xl border border-white/10 p-6 rounded-2xl grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="space-y-4">
            <div className="space-y-2">
              <label className="text-sm font-medium text-gray-400 flex items-center gap-2">
                <User className="w-4 h-4" />{" "}
                {uiStrings?.label_customer_name || "Customer Name"}
                <span className="text-red-500">*</span>
              </label>
              <div className={showErrors && isFieldMissing(formData.customer_name) ? "ring-1 ring-red-500 rounded-lg" : ""}>
                <CustomerSearch 
                  defaultValue={formData.customer_name}
                  onSelect={handleCustomerSelect}
                />
              </div>
            </div>
            
            <div className="grid grid-cols-2 gap-4">
               <div className="space-y-2">
                <label className="text-sm font-medium text-gray-400 flex items-center gap-2">
                  <Phone className="w-4 h-4" /> Mobile Number
                  <span className="text-red-500">*</span>
                </label>
                <input
                  type="text"
                  value={formData?.customer_phone || ""}
                  onChange={(e) => setFormData(prev => ({ ...prev, customer_phone: e.target.value }))}
                  onBlur={(e) => {
                    let val = e.target.value.replace(/\s/g, "");
                    if (/^\d{10}$/.test(val)) {
                      setFormData(prev => ({ ...prev, customer_phone: `+91 ${val.substring(0, 5)} ${val.substring(5)}` }));
                    }
                  }}
                  className={`w-full bg-white/5 border rounded-lg px-4 py-2 text-white focus:outline-none focus:border-primary transition-colors ${
                    showErrors && isFieldMissing(formData.customer_phone) ? "border-red-500" : "border-white/10"
                  }`}
                />
              </div>
              <div className="space-y-2">
                <label className="text-sm font-medium text-gray-400 flex items-center gap-2">
                  <Mail className="w-4 h-4" /> Email (Optional)
                </label>
                <input
                  type="email"
                  value={formData?.customer_email || ""}
                  onChange={(e) => setFormData({ ...formData, customer_email: e.target.value })}
                  className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-primary transition-colors"
                />
              </div>
            </div>

            <div className="space-y-2">
              <label className="text-sm font-medium text-gray-400 flex items-center gap-2">
                <MapPin className="w-4 h-4" /> Billing Address
                <span className="text-red-500">*</span>
              </label>
              <textarea
                value={formData?.billing_address || ""}
                onChange={(e) => setFormData(prev => ({ ...prev, billing_address: e.target.value }))}
                className={`w-full bg-white/5 border rounded-lg px-4 py-2 text-white text-sm focus:outline-none focus:border-primary transition-colors h-20 resize-none ${
                  showErrors && isFieldMissing(formData.billing_address) ? "border-red-500" : "border-white/10"
                }`}
              />
            </div>
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium text-gray-400 flex items-center gap-2">
                <Calendar className="w-4 h-4" />{" "}
                {uiStrings?.label_invoice_date || "Invoice Date"}
                <span className="text-red-500">*</span>
              </label>
              <input
                type="date"
                value={formData.invoice_date?.split("T")[0]}
                onChange={(e) =>
                  setFormData({ ...formData, invoice_date: e.target.value })
                }
                className={`w-full bg-white/5 border rounded-lg px-4 py-2 text-white focus:outline-none focus:border-primary transition-colors ${
                  showErrors && isFieldMissing(formData.invoice_date) ? "border-red-500" : "border-white/10"
                }`}
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium text-gray-400 flex items-center gap-2">
                <Calendar className="w-4 h-4" />{" "}
                {uiStrings?.label_due_date || "Due Date"}
                <span className="text-red-500">*</span>
              </label>
              <input
                type="date"
                value={formData.due_date?.split("T")[0]}
                onChange={(e) =>
                  setFormData({ ...formData, due_date: e.target.value })
                }
                className={`w-full bg-white/5 border rounded-lg px-4 py-2 text-white focus:outline-none focus:border-primary transition-colors ${
                  showErrors && isFieldMissing(formData.due_date) ? "border-red-500" : "border-white/10"
                }`}
              />
            </div>
            <div className="col-span-2 space-y-2 pt-4 border-t border-white/5">
                <label className="text-xs font-medium text-gray-500 uppercase">Currency</label>
                <div className="text-lg font-bold text-white flex items-center gap-2">
                   <IndianRupee className="w-5 h-5 text-primary" /> INR (Standardized)
                </div>
            </div>
          </div>
        </div>

        {/* Transporter & Bank Details */}
        <div className="bg-gray-900/50 backdrop-blur-xl border border-white/10 p-6 rounded-2xl space-y-6">
          <h3 className="text-md font-semibold text-white flex items-center gap-2 border-b border-white/5 pb-3">
             <IndianRupee className="w-4 h-4 text-primary" /> Transporter Credentials & Bank
          </h3>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div className="space-y-2">
              <label className="text-xs font-medium text-gray-500 uppercase">GSTIN (Seller)</label>
              <input
                type="text"
                value={formData.gstin_seller}
                onChange={(e) => setFormData(prev => ({ ...prev, gstin_seller: e.target.value.toUpperCase() }))}
                placeholder="29AAAAA0000A1Z5"
                className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-2 text-white focus:border-primary transition-colors"
              />
            </div>
            <div className="space-y-2">
              <label className="text-xs font-medium text-gray-500 uppercase">PAN Number</label>
              <input
                type="text"
                value={formData.pan_seller}
                onChange={(e) => setFormData(prev => ({ ...prev, pan_seller: e.target.value.toUpperCase() }))}
                placeholder="ABCDE1234F"
                className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-2 text-white focus:border-primary transition-colors"
              />
            </div>
            <div className="space-y-2">
              <label className="text-xs font-medium text-gray-500 uppercase">Place of Supply</label>
              <select
                 value={formData.place_of_supply}
                 onChange={(e) => setFormData(prev => ({ ...prev, place_of_supply: e.target.value }))}
                 className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-2 text-white focus:border-primary transition-colors"
              >
                <option value="KA">Karnataka (29)</option>
                <option value="MH">Maharashtra (27)</option>
                <option value="DL">Delhi (07)</option>
                <option value="TN">Tamil Nadu (33)</option>
                <option value="TS">Telangana (36)</option>
              </select>
            </div>
          </div>
          
          <div className="grid grid-cols-1 md:grid-cols-4 gap-4 pt-2">
             <div className="space-y-2">
                <label className="text-xs font-medium text-gray-500 uppercase">Bank Name</label>
                <input
                  type="text"
                  value={formData.bank_name}
                  onChange={(e) => setFormData(prev => ({ ...prev, bank_name: e.target.value }))}
                  className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-2 text-white text-sm"
                />
             </div>
             <div className="space-y-2">
                <label className="text-xs font-medium text-gray-500 uppercase">A/C Number</label>
                <input
                  type="text"
                  value={formData.account_no}
                  onChange={(e) => setFormData(prev => ({ ...prev, account_no: e.target.value }))}
                  className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-2 text-white text-sm"
                />
             </div>
             <div className="space-y-2">
                <label className="text-xs font-medium text-gray-500 uppercase">IFSC Code</label>
                <input
                  type="text"
                  value={formData.ifsc_code}
                  onChange={(e) => setFormData(prev => ({ ...prev, ifsc_code: e.target.value.toUpperCase() }))}
                  className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-2 text-white text-sm"
                />
             </div>
             <div className="space-y-2">
                <label className="text-xs font-medium text-gray-500 uppercase">Beneficiary</label>
                <input
                  type="text"
                  value={formData.beneficiary_name}
                  onChange={(e) => setFormData(prev => ({ ...prev, beneficiary_name: e.target.value }))}
                  className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-2 text-white text-sm"
                />
             </div>
          </div>
        </div>

        {/* Logistics & Trip Details */}
        <div className="bg-gray-900/50 backdrop-blur-xl border border-white/10 p-6 rounded-2xl space-y-6">
          <h3 className="text-md font-semibold text-white flex items-center gap-2 border-b border-white/5 pb-3">
             <MapPin className="w-4 h-4 text-primary" /> Logistics & Consignment
          </h3>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
             <div className="space-y-2">
                <label className="text-xs font-medium text-gray-500 uppercase">Vehicle Register No.</label>
                <input
                  type="text"
                  value={formData?.vehicle_no || ""}
                  onChange={(e) => setFormData(prev => ({ ...prev, vehicle_no: e.target.value.toUpperCase() }))}
                  placeholder="KA-01-AB-1234"
                  className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-2 text-white focus:border-primary transition-colors"
                />
             </div>
             <div className="space-y-2">
                <label className="text-xs font-medium text-gray-500 uppercase">E-Way Bill No.</label>
                <input
                  type="text"
                  value={formData?.eway_bill_no || ""}
                  onChange={(e) => setFormData(prev => ({ ...prev, eway_bill_no: e.target.value }))}
                  className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-2 text-white focus:border-primary transition-colors"
                />
             </div>
             <div className="space-y-2">
                <label className="text-xs font-medium text-gray-500 uppercase">Billed Distance (KM)</label>
                <input
                  type="number"
                  value={formData?.billed_distance || 0}
                  onChange={(e) => setFormData(prev => ({ ...prev, billed_distance: parseFloat(e.target.value) }))}
                  className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-2 text-white focus:border-primary transition-colors"
                />
             </div>
          </div>
          
          <div className="grid grid-cols-1 md:grid-cols-2 gap-6 pt-2">
             <div className="space-y-2">
                <label className="text-xs font-medium text-gray-500 uppercase">Pickup Location</label>
                <input
                  type="text"
                  value={formData?.pickup_location || ""}
                  onChange={(e) => setFormData(prev => ({ ...prev, pickup_location: e.target.value }))}
                  className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-2 text-white text-sm"
                />
             </div>
             <div className="space-y-2">
                <label className="text-xs font-medium text-gray-500 uppercase">Drop Location</label>
                <input
                  type="text"
                  value={formData?.drop_location || ""}
                  onChange={(e) => setFormData(prev => ({ ...prev, drop_location: e.target.value }))}
                  className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-2 text-white text-sm"
                />
             </div>
          </div>
        </div>

        {/* Line Items Section */}
        <div className="bg-gray-900/50 backdrop-blur-xl border border-white/10 p-6 rounded-2xl space-y-4">
          <div className="flex items-center justify-between">
            <h3 className="text-lg font-semibold text-white flex items-center gap-2">
              <Package className="w-5 h-5" />{" "}
              {uiStrings?.title_line_items || "Line Items"}
            </h3>
            <button
              type="button"
              onClick={handleAddItem}
              className="flex items-center gap-2 text-primary hover:text-primary/80 transition-colors text-sm font-medium"
            >
              <Plus className="w-4 h-4" />
              {uiStrings?.btn_add_item || "Add Item"}
            </button>
          </div>

          <div className="space-y-3">
            {formData.items?.map((item, index) => (
              <div
                key={index}
                className="grid grid-cols-12 gap-3 p-4 bg-white/5 rounded-xl border border-white/10 items-end"
              >
                <div className="col-span-12 md:col-span-5 space-y-2">
                  <label className="text-xs font-medium text-gray-500 uppercase tracking-wider">
                    {uiStrings?.th_desc || "Description"}
                    <span className="text-red-500 ml-1">*</span>
                  </label>
                  <input
                    type="text"
                    value={item.description}
                    onChange={(e) =>
                      handleItemChange(index, "description", e.target.value)
                    }
                    className={`w-full bg-black/20 border rounded-lg px-3 py-1.5 text-white text-sm focus:outline-none focus:border-primary/50 ${
                      showErrors && isFieldMissing(item.description) ? "border-red-500" : "border-white/5"
                    }`}
                    placeholder="e.g. Full Load Transport"
                  />
                </div>
                <div className="col-span-12 md:col-span-2 space-y-2">
                  <label className="text-xs font-medium text-gray-500 uppercase tracking-wider">
                    HSN/SAC
                  </label>
                  <input
                    type="text"
                    value={item.hsn_sac}
                    onChange={(e) =>
                      handleItemChange(index, "hsn_sac", e.target.value)
                    }
                    className="w-full bg-black/20 border border-white/5 rounded-lg px-3 py-1.5 text-white text-sm focus:outline-none focus:border-primary/50"
                  />
                </div>
                <div className="col-span-4 md:col-span-1 space-y-2">
                  <label className="text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Qty
                    <span className="text-red-500 ml-1">*</span>
                  </label>
                  <input
                    type="number"
                    min="1"
                    value={item.quantity}
                    onChange={(e) =>
                      handleItemChange(
                        index,
                        "quantity",
                        parseFloat(e.target.value) || 0,
                      )
                    }
                    className={`w-full bg-black/20 border rounded-lg px-3 py-1.5 text-white text-sm focus:outline-none focus:border-primary/50 ${
                      showErrors && item.quantity <= 0 ? "border-red-500" : "border-white/5"
                    }`}
                  />
                </div>
                <div className="col-span-4 md:col-span-1 space-y-2">
                  <label className="text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Unit
                  </label>
                   <select
                    value={item.unit}
                    onChange={(e) =>
                      handleItemChange(index, "unit", e.target.value)
                    }
                    className="w-full bg-black/20 border border-white/5 rounded-lg px-2 py-1.5 text-white text-sm focus:outline-none focus:border-primary/50"
                  >
                    <option value="Trip">Trip</option>
                    <option value="Box">Box</option>
                    <option value="Ton">Ton</option>
                    <option value="KM">KM</option>
                  </select>
                </div>
                <div className="col-span-4 md:col-span-2 space-y-2">
                  <label className="text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Rate
                    <span className="text-red-500 ml-1">*</span>
                  </label>
                  <div className={`relative border rounded-lg ${showErrors && item.unit_price <= 0 ? "border-red-500" : "border-transparent"}`}>
                    <IndianRupee className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3 h-3 text-gray-500" />
                    <input
                      type="number"
                      value={item.unit_price}
                      onChange={(e) =>
                        handleItemChange(
                          index,
                          "unit_price",
                          parseFloat(e.target.value) || 0,
                        )
                      }
                      className="w-full bg-black/20 border border-white/5 rounded-lg pl-7 pr-3 py-1.5 text-white text-sm focus:outline-none focus:border-primary/50"
                    />
                  </div>
                </div>
                <div className="col-span-3 md:col-span-2 space-y-2">
                  <label className="text-xs font-medium text-gray-500 uppercase tracking-wider">
                    {uiStrings?.th_tax || "Tax"} (%)
                  </label>
                  <input
                    required
                    type="number"
                    value={item.tax_percent}
                    onChange={(e) =>
                      handleItemChange(
                        index,
                        "tax_percent",
                        parseFloat(e.target.value) || 0,
                      )
                    }
                    className="w-full bg-black/20 border border-white/5 rounded-lg px-3 py-1.5 text-white text-sm focus:outline-none focus:border-primary/50"
                  />
                </div>
                <div className="col-span-1 flex justify-end pb-1.5">
                  <button
                    type="button"
                    onClick={() => handleRemoveItem(index)}
                    disabled={formData.items!.length <= 1}
                    className="p-1.5 text-gray-500 hover:text-red-500 transition-colors disabled:opacity-30"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                </div>
              </div>
            ))}
          </div>
        </div>

        {/* Totals Section */}
        <div className="flex flex-col md:flex-row gap-6">
          <div className="flex-1 bg-gray-900/50 backdrop-blur-xl border border-white/10 p-6 rounded-2xl">
            <label className="text-sm font-medium text-gray-400 block mb-2">
              Notes / Terms
            </label>
            <textarea
              value={formData.notes || ""}
              onChange={(e) => setFormData({ ...formData, notes: e.target.value })}
              className="w-full h-24 bg-white/5 border border-white/10 rounded-lg px-4 py-2 text-white text-sm focus:outline-none focus:border-primary transition-colors resize-none"
              placeholder="Payment is due within 14 days..."
            ></textarea>
          </div>
          <div className="w-full md:w-80 bg-gray-900/50 backdrop-blur-xl border border-white/10 p-6 rounded-2xl space-y-3">
            <div className="flex justify-between text-gray-400 text-sm">
              <span>Subtotal</span>
              <span>₹{calculateSubtotal().toLocaleString()}</span>
            </div>
            <div className="flex justify-between text-gray-400 text-sm">
              <span>Tax</span>
              <span>₹{calculateTax().toLocaleString()}</span>
            </div>
            <div className="pt-3 border-t border-white/10 flex justify-between text-white font-bold text-lg">
              <span>Total</span>
              <span>₹{calculateTotal().toLocaleString('en-IN')}</span>
            </div>
          </div>
        </div>

        {/* Action Bar */}
        <div className="flex items-center justify-end gap-3">
          <button
            type="button"
            onClick={() => navigate(-1)}
            className="px-6 py-2.5 bg-gray-800 text-gray-300 rounded-xl hover:bg-gray-700 transition-colors font-medium"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={loading}
            className="flex items-center gap-2 px-8 py-2.5 bg-primary text-white rounded-xl hover:bg-primary/90 transition-all font-semibold shadow-lg shadow-primary/20 disabled:opacity-50"
          >
            {loading ? (
              <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white"></div>
            ) : (
              <Save className="w-4 h-4" />
            )}
            {isEdit ? "Update Invoice" : "Issue Invoice"}
          </button>
        </div>
      </form>
    </div>
  );
};

export default InvoiceForm;
