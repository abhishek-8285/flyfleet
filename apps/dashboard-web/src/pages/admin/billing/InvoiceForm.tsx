import {
    Calendar,
    ChevronLeft,
    IndianRupee,
    Package,
    Plus,
    Save,
    Trash2,
    User,
} from "lucide-react";
import React, { useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useUIStrings } from "../../../hooks/useUIStrings";
import billingService from "../../../services/billingService";
import { Invoice, InvoiceItem, InvoiceStatus } from "../../../types/billing";

const InvoiceForm: React.FC = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const isEdit = !!id;

  const { uiStrings } = useUIStrings("admin_billing");

  const [formData, setFormData] = useState<Partial<Invoice>>({
    customer_name: "",
    invoice_date: new Date().toISOString().split("T")[0],
    due_date: new Date(Date.now() + 14 * 24 * 60 * 60 * 1000)
      .toISOString()
      .split("T")[0],
    currency: "INR",
    items: [
      {
        id: "",
        invoice_id: "",
        description: "",
        quantity: 1,
        unit_price: 0,
        tax_percent: 18,
        total: 0,
      },
    ],
  });

  const [loading, setLoading] = useState(false);
  const [billingMeta, setBillingMeta] = useState<any>(null);

  useEffect(() => {
    fetch("/api/v1/ui/global_config")
      .then((res) => res.json())
      .then((data) => {
        setBillingMeta(data.billing_metadata);
        if (!isEdit && data.billing_metadata) {
          setFormData((prev) => ({
            ...prev,
            currency:
              data.billing_metadata.currency_symbol === "₹" ? "INR" : "USD",
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

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setLoading(true);
    try {
      const subtotal = calculateSubtotal();
      const tax = calculateTax();
      const total = calculateTotal();

      const payload = {
        customer_name: formData.customer_name,
        invoice_date: formData.invoice_date,
        due_date: formData.due_date,
        currency: formData.currency || "INR",
        status: "PENDING" as InvoiceStatus,
        subtotal,
        tax_amount: tax,
        total_amount: total,
        amount_paid: 0,
        balance_due: total,
        notes: formData.notes || "",
        items: (formData.items || []).map((item) => ({
          description: item.description,
          quantity: item.quantity,
          unit_price: item.unit_price,
          tax_percent: item.tax_percent,
          total: item.quantity * item.unit_price,
        })),
      };

      if (isEdit) {
        await billingService.updateInvoice(id!, payload as Partial<Invoice>);
      } else {
        await billingService.createInvoice(payload as Partial<Invoice>);
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
        <div className="w-20"></div> {/* Spacer */}
      </div>

      <form onSubmit={handleSubmit} className="space-y-6">
        {/* Customer & Dates Section */}
        <div className="bg-gray-900/50 backdrop-blur-xl border border-white/10 p-6 rounded-2xl grid grid-cols-1 md:grid-cols-2 gap-6">
          <div className="space-y-2">
            <label className="text-sm font-medium text-gray-400 flex items-center gap-2">
              <User className="w-4 h-4" />{" "}
              {uiStrings?.label_customer_name || "Customer Name"}
            </label>
            <input
              required
              type="text"
              value={formData.customer_name}
              onChange={(e) =>
                setFormData({ ...formData, customer_name: e.target.value })
              }
              className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-primary transition-colors"
              placeholder={uiStrings?.ph_customer_name || "e.g. Acme Corp"}
            />
          </div>
          <div className="grid grid-cols-2 gap-4">
            <div className="space-y-2">
              <label className="text-sm font-medium text-gray-400 flex items-center gap-2">
                <Calendar className="w-4 h-4" />{" "}
                {uiStrings?.label_invoice_date || "Invoice Date"}
              </label>
              <input
                required
                type="date"
                value={formData.invoice_date?.split("T")[0]}
                onChange={(e) =>
                  setFormData({ ...formData, invoice_date: e.target.value })
                }
                className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-primary transition-colors"
              />
            </div>
            <div className="space-y-2">
              <label className="text-sm font-medium text-gray-400 flex items-center gap-2">
                <Calendar className="w-4 h-4" />{" "}
                {uiStrings?.label_due_date || "Due Date"}
              </label>
              <input
                required
                type="date"
                value={formData.due_date?.split("T")[0]}
                onChange={(e) =>
                  setFormData({ ...formData, due_date: e.target.value })
                }
                className="w-full bg-white/5 border border-white/10 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-primary transition-colors"
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
                  </label>
                  <input
                    required
                    type="text"
                    value={item.description}
                    onChange={(e) =>
                      handleItemChange(index, "description", e.target.value)
                    }
                    className="w-full bg-black/20 border border-white/5 rounded-lg px-3 py-1.5 text-white text-sm focus:outline-none focus:border-primary/50"
                    placeholder={
                      uiStrings?.ph_item_desc || "Describe service or product"
                    }
                  />
                </div>
                <div className="col-span-4 md:col-span-2 space-y-2">
                  <label className="text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Qty
                  </label>
                  <input
                    required
                    type="number"
                    min="1"
                    value={item.quantity}
                    onChange={(e) =>
                      handleItemChange(
                        index,
                        "quantity",
                        parseInt(e.target.value) || 0,
                      )
                    }
                    className="w-full bg-black/20 border border-white/5 rounded-lg px-3 py-1.5 text-white text-sm focus:outline-none focus:border-primary/50"
                  />
                </div>
                <div className="col-span-4 md:col-span-2 space-y-2">
                  <label className="text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Price
                  </label>
                  <div className="relative">
                    <IndianRupee className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3 h-3 text-gray-500" />
                    <input
                      required
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
              <span>₹{calculateTotal().toLocaleString()}</span>
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
