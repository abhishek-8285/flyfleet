import {
  ArrowLeftIcon,
  BanknotesIcon,
  BuildingOffice2Icon,
  ClipboardDocumentCheckIcon,
  ClockIcon,
  EnvelopeIcon,
  PencilIcon,
  PhoneIcon,
  ShieldCheckIcon,
  TruckIcon,
} from "@heroicons/react/24/outline";
import React, { useEffect, useState } from "react";
import { Link, useNavigate, useParams } from "react-router-dom";
import {
  customerV2Service,
  Customer,
  CustomerSummary,
} from "../../services/customerV2Service";
import { EditCustomerModal } from "./EditCustomerModal";
import { LoadingSpinner } from "../../components/common/LoadingSpinner";

const EMPTY_SUMMARY: CustomerSummary = {
  active_shipments: 0,
  total_shipments: 0,
  pending_bookings: 0,
  total_revenue: 0,
  unpaid_invoices: 0,
};

function formatDate(value?: string) {
  if (!value) return "Not recorded";
  const parsed = new Date(value);
  if (Number.isNaN(parsed.getTime())) return "Not recorded";
  return parsed.toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
  });
}

function formatCurrency(value: number) {
  return `\u20b9${value.toLocaleString()}`;
}

export const CustomerDetail: React.FC = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const [customer, setCustomer] = useState<Customer | null>(null);
  const [summary, setSummary] = useState<CustomerSummary>(EMPTY_SUMMARY);
  const [loading, setLoading] = useState(true);
  const [summaryUnavailable, setSummaryUnavailable] = useState(false);
  const [activeTab, setActiveTab] = useState("shipments");
  const [isEditModalOpen, setIsEditModalOpen] = useState(false);

  useEffect(() => {
    if (!id) return;

    const fetchData = async () => {
      try {
        setLoading(true);
        setSummaryUnavailable(false);

        const customerData = await customerV2Service.getCustomerById(id);
        setCustomer(customerData);

        try {
          const summaryData = await customerV2Service.getCustomerSummary(id);
          setSummary(summaryData);
        } catch (error: any) {
          if (error?.response?.status === 404) {
            setSummary(EMPTY_SUMMARY);
            setSummaryUnavailable(true);
          } else {
            throw error;
          }
        }
      } catch (error) {
        console.error("Failed to fetch customer detail:", error);
        setCustomer(null);
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [id]);

  if (loading) {
    return (
      <div className="flex flex-col items-center justify-center py-24 gap-4 bg-surface-dim min-h-screen">
        <div className="animate-spin rounded-full h-10 w-10 border-b-2 border-primary" />
        <p className="text-on-surface-variant font-black uppercase tracking-[0.2em] text-[10px]">
          Initialising Customer Profile...
        </p>
      </div>
    );
  }

  const handleEditSuccess = (updated: Customer) => {
    setCustomer(updated);
  };

  if (!customer) {
    return (
      <div className="flex flex-col items-center justify-center py-24 text-on-surface-variant bg-surface-dim min-h-screen text-center">
        <span className="material-symbols-outlined text-6xl text-rose-500/20 mb-4">
          corporate_fare
        </span>
        <h2 className="text-xl font-bold text-white mb-2">
          Customer Record Not Found
        </h2>
        <p className="mb-6 text-sm">
          The requested customer profile could not be retrieved from the
          registry.
        </p>
        <button
          onClick={() => navigate("/admin/customers")}
          className="px-6 py-2.5 bg-primary text-white rounded-xl hover:bg-primary-dark transition-all font-bold uppercase text-[10px] tracking-widest shadow-lg shadow-primary/20 active:scale-95"
        >
          Return to directory
        </button>
      </div>
    );
  }

  const record = customer as Customer & Record<string, any>;
  const contactEmail = record.contact_email || customer.email || "Not configured";
  const contactPhone = record.contact_phone || customer.phone || "Not configured";
  const billingAddress =
    record.billing_address || customer.address || "Billing address not configured.";
  const gstNumber = record.gst_number || customer.gstin;
  const createdAt = record.created_at || customer.created_at;
  const updatedAt = record.updated_at || customer.updated_at;
  const status = customer.status || "ACTIVE";

  const stats = [
    {
      label: "Active Shipments",
      value: `${summary.active_shipments}`,
      icon: TruckIcon,
      tone: "text-primary bg-primary/10 border-primary/20",
    },
    {
      label: "Pending Bookings",
      value: `${summary.pending_bookings}`,
      icon: ClockIcon,
      tone: "text-amber-400 bg-amber-500/10 border-amber-500/20",
    },
    {
      label: "Total Shipments",
      value: `${summary.total_shipments}`,
      icon: ClipboardDocumentCheckIcon,
      tone: "text-emerald-400 bg-emerald-500/10 border-emerald-500/20",
    },
    {
      label: "Unpaid Invoices",
      value: `${summary.unpaid_invoices}`,
      icon: BanknotesIcon,
      tone: "text-rose-400 bg-rose-500/10 border-rose-500/20",
    },
    {
      label: "Revenue",
      value: formatCurrency(summary.total_revenue),
      icon: BanknotesIcon,
      tone: "text-sky-400 bg-sky-500/10 border-sky-500/20",
    },
  ];

  const tabs = [
    {
      id: "shipments",
      label: "Active Shipments",
      icon: TruckIcon,
      title: "Operational movement sync pending",
      description:
        "Shipment-level movement data will appear here once the admin customer feeds are wired into the detail surface.",
    },
    {
      id: "history",
      label: "Booking History",
      icon: ClipboardDocumentCheckIcon,
      title: "Booking history not connected yet",
      description:
        "Historical request and confirmation records are not yet exposed on the admin customer detail route.",
    },
    {
      id: "audit",
      label: "Security & Audit",
      icon: ShieldCheckIcon,
      title: "Audit timeline not connected yet",
      description:
        "Access, change, and policy events will surface here after the customer audit stream is added.",
    },
  ];

  const activeTabConfig = tabs.find((tab) => tab.id === activeTab) || tabs[0];

  return (
    <div className="flex-1 flex flex-col bg-surface-dim p-8 -mx-6 -mt-6 gap-8 min-h-screen">
      <div className="flex flex-col md:flex-row justify-between items-start md:items-end gap-6 bg-surface-container border border-outline rounded-3xl p-8 shadow-2xl">
        <div>
          <nav className="flex items-center gap-2 text-[10px] font-black uppercase tracking-widest text-on-surface-variant mb-4">
            <Link
              to="/admin/customers"
              className="hover:text-primary transition-colors"
            >
              Customer Directory
            </Link>
            <span className="material-symbols-outlined text-[10px]">
              chevron_right
            </span>
            <span className="text-white font-mono">#{customer.id.slice(0, 8)}</span>
          </nav>
          <div className="flex items-start gap-6">
            <button
              onClick={() => navigate("/admin/customers")}
              className="h-12 w-12 flex items-center justify-center bg-surface-dim border border-outline rounded-xl hover:bg-surface-container-highest transition-all group"
            >
              <ArrowLeftIcon className="h-5 w-5 text-on-surface-variant group-hover:-translate-x-1 transition-transform" />
            </button>
            <div>
              <div className="flex flex-wrap items-center gap-3 mb-3">
                <h1 className="text-4xl font-bold text-white tracking-tighter">
                  {customer.name}
                </h1>
                <span
                  className={`px-3 py-1 rounded-lg text-[10px] font-black uppercase tracking-widest border ${
                    status === "ACTIVE"
                      ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
                      : status === "PENDING"
                        ? "bg-amber-500/10 text-amber-400 border-amber-500/20"
                        : "bg-surface-dim text-on-surface-variant border-outline"
                  }`}
                >
                  {status}
                </span>
                {gstNumber && (
                  <span className="px-3 py-1 rounded-lg text-[10px] font-black uppercase tracking-widest border bg-primary/10 text-primary border-primary/20">
                    GST Registered
                  </span>
                )}
              </div>
              <p className="text-on-surface-variant text-[11px] font-bold uppercase tracking-wider flex flex-wrap items-center gap-2">
                <span className="material-symbols-outlined text-sm">
                  corporate_fare
                </span>
                Customer ID:
                <span className="text-white font-mono">{customer.id}</span>
                <span className="w-1 h-1 rounded-full bg-outline mx-2" />
                <span className="material-symbols-outlined text-sm">event</span>
                Created {formatDate(createdAt)}
              </p>
            </div>
          </div>
        </div>
        <div className="flex gap-3">
          <button className="bg-surface-dim border border-outline text-white px-6 py-3 rounded-xl text-[10px] font-black uppercase tracking-widest hover:bg-surface-container-highest transition-all flex items-center gap-3 active:scale-95 shadow-lg">
            <span className="material-symbols-outlined text-base">sync</span>
            Refresh profile
          </button>
          <button 
            onClick={() => setIsEditModalOpen(true)}
            className="bg-primary text-white px-6 py-3 rounded-xl text-[10px] font-black uppercase tracking-widest shadow-lg shadow-primary/20 hover:bg-primary-dark transition-all flex items-center gap-3 active:scale-95"
          >
            <PencilIcon className="h-4 w-4" />
            Edit customer
          </button>
        </div>
      </div>

      {summaryUnavailable && (
        <div className="border border-amber-500/20 bg-amber-500/10 rounded-2xl px-5 py-4 text-sm text-amber-300">
          Summary metrics are not available from the backend yet. The customer
          record is loaded, but shipment and revenue cards are showing fallback
          values.
        </div>
      )}

      <div className="grid grid-cols-1 lg:grid-cols-12 gap-8">
        <div className="lg:col-span-8 flex flex-col gap-8">
          <div className="bg-surface-container rounded-[2.5rem] border border-outline relative overflow-hidden shadow-2xl">
            <div className="absolute inset-0 bg-gradient-to-br from-primary/10 via-transparent to-transparent pointer-events-none" />
            <div className="relative z-10 p-8 md:p-10">
              <div className="flex items-center gap-3 mb-8">
                <div className="h-12 w-12 bg-primary/10 rounded-xl flex items-center justify-center border border-primary/20">
                  <BuildingOffice2Icon className="h-6 w-6 text-primary" />
                </div>
                <div>
                  <p className="text-[10px] font-black text-on-surface-variant tracking-[0.2em] uppercase mb-1">
                    Customer Registry
                  </p>
                  <h2 className="text-2xl font-bold text-white tracking-tight">
                    Account overview
                  </h2>
                </div>
              </div>

              <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                <div className="bg-surface-dim/80 border border-outline rounded-2xl p-6">
                  <p className="text-[10px] font-black text-on-surface-variant uppercase tracking-[0.18em] mb-2">
                    Primary Email
                  </p>
                  <p className="text-base font-bold text-white break-all flex items-center gap-3">
                    <EnvelopeIcon className="h-5 w-5 text-primary" />
                    {contactEmail}
                  </p>
                </div>
                <div className="bg-surface-dim/80 border border-outline rounded-2xl p-6">
                  <p className="text-[10px] font-black text-on-surface-variant uppercase tracking-[0.18em] mb-2">
                    Phone Number
                  </p>
                  <p className="text-base font-bold text-white flex items-center gap-3">
                    <PhoneIcon className="h-5 w-5 text-primary" />
                    {contactPhone}
                  </p>
                </div>
                <div className="bg-surface-dim/80 border border-outline rounded-2xl p-6 md:col-span-2">
                  <p className="text-[10px] font-black text-on-surface-variant uppercase tracking-[0.18em] mb-2">
                    Billing Address
                  </p>
                  <p className="text-sm font-bold text-white leading-relaxed">
                    {billingAddress}
                  </p>
                </div>
              </div>
            </div>
          </div>

          <div className="grid grid-cols-2 md:grid-cols-3 xl:grid-cols-5 gap-6">
            {stats.map((stat) => (
              <div
                key={stat.label}
                className="bg-surface-container border border-outline p-6 rounded-[2rem] shadow-xl"
              >
                <div className="flex items-center justify-between mb-6">
                  <div className={`h-11 w-11 rounded-xl border flex items-center justify-center ${stat.tone}`}>
                    <stat.icon className="h-5 w-5" />
                  </div>
                  <span className="text-[9px] font-black text-on-surface-variant uppercase tracking-[0.2em]">
                    Live
                  </span>
                </div>
                <div className="text-3xl font-bold text-white tracking-tight">
                  {stat.value}
                </div>
                <div className="text-[10px] font-black text-on-surface-variant uppercase tracking-widest mt-2">
                  {stat.label}
                </div>
              </div>
            ))}
          </div>
        </div>

        <div className="lg:col-span-4 flex flex-col gap-8">
          <div className="bg-surface-container rounded-[2.5rem] border border-outline shadow-2xl overflow-hidden">
            <div className="p-8 border-b border-outline bg-surface-container/50">
              <h3 className="text-[11px] font-black text-white tracking-[0.2em] uppercase flex items-center gap-3">
                <ShieldCheckIcon className="h-5 w-5 text-primary" />
                Account Signals
              </h3>
            </div>
            <div className="p-8 space-y-5">
              <div className="bg-surface-dim border border-outline rounded-2xl p-5">
                <p className="text-[10px] font-black text-on-surface-variant uppercase tracking-widest mb-2">
                  Compliance Status
                </p>
                <p className="text-sm font-bold text-white">
                  {gstNumber ? "GST registered" : "Tax registration missing"}
                </p>
              </div>
              <div className="bg-surface-dim border border-outline rounded-2xl p-5">
                <p className="text-[10px] font-black text-on-surface-variant uppercase tracking-widest mb-2">
                  Last Updated
                </p>
                <p className="text-sm font-bold text-white">
                  {formatDate(updatedAt)}
                </p>
              </div>
              <div className="bg-surface-dim border border-outline rounded-2xl p-5">
                <p className="text-[10px] font-black text-on-surface-variant uppercase tracking-widest mb-2">
                  Registry Identifier
                </p>
                <p className="text-sm font-mono font-bold text-white break-all">
                  {customer.id}
                </p>
              </div>
            </div>
          </div>

          <div className="bg-surface-container rounded-[2.5rem] border border-outline shadow-2xl overflow-hidden">
            <div className="p-8 border-b border-outline bg-surface-container/50">
              <h3 className="text-[11px] font-black text-white tracking-[0.2em] uppercase flex items-center gap-3">
                <BanknotesIcon className="h-5 w-5 text-primary" />
                Commercial Profile
              </h3>
            </div>
            <div className="p-8">
              <p className="text-[10px] font-black text-on-surface-variant uppercase tracking-[0.18em] mb-2">
                Tracked Revenue
              </p>
              <div className="text-4xl font-bold text-white tracking-tighter mb-6">
                {formatCurrency(summary.total_revenue)}
              </div>
              <div className="space-y-4">
                <div className="flex items-center justify-between text-sm">
                  <span className="text-on-surface-variant font-bold">
                    Open invoice count
                  </span>
                  <span className="text-white font-black">
                    {summary.unpaid_invoices}
                  </span>
                </div>
                <div className="flex items-center justify-between text-sm">
                  <span className="text-on-surface-variant font-bold">
                    Shipment throughput
                  </span>
                  <span className="text-white font-black">
                    {summary.total_shipments}
                  </span>
                </div>
                <div className="flex items-center justify-between text-sm">
                  <span className="text-on-surface-variant font-bold">
                    Pending bookings
                  </span>
                  <span className="text-white font-black">
                    {summary.pending_bookings}
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div className="bg-surface-container rounded-[2.5rem] border border-outline shadow-2xl overflow-hidden min-h-[360px]">
        <div className="border-b border-outline px-8 py-2 flex flex-wrap items-center gap-8">
          {tabs.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`flex items-center gap-2 py-5 text-[10px] font-black uppercase tracking-widest transition-all border-b-2 ${
                activeTab === tab.id
                  ? "border-primary text-white"
                  : "border-transparent text-on-surface-variant hover:text-white"
              }`}
            >
              <tab.icon className="h-4 w-4" />
              <span>{tab.label}</span>
            </button>
          ))}
        </div>
        <div className="p-12 md:p-16 text-center">
          <div className="inline-flex h-16 w-16 items-center justify-center rounded-2xl bg-surface-dim border border-outline mb-6">
            <activeTabConfig.icon className="h-8 w-8 text-primary" />
          </div>
          <h3 className="text-xl font-bold text-white tracking-tight mb-3">
            {activeTabConfig.title}
          </h3>
          <p className="text-on-surface-variant text-sm max-w-2xl mx-auto leading-relaxed">
            {activeTabConfig.description}
          </p>
        </div>
      </div>

      <EditCustomerModal 
        isOpen={isEditModalOpen}
        onClose={() => setIsEditModalOpen(false)}
        customer={customer}
        onSuccess={handleEditSuccess}
      />
    </div>
  );
};
