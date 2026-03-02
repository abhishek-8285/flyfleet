import React, { useState } from "react";
import { useUIStrings } from "../../hooks/useUIStrings";

interface AddVehicleModalProps {
  isOpen: boolean;
  onClose: () => void;
}

export const AddVehicleModal: React.FC<AddVehicleModalProps> = ({
  isOpen,
  onClose,
}) => {
  const [step, setStep] = useState(1);
  const [vin, setVin] = useState("");
  const [isValidatingVin, setIsValidatingVin] = useState(false);
  const [vinError, setVinError] = useState("");
  const [vinSuccess, setVinSuccess] = useState(false);
  const { uiStrings } = useUIStrings("admin");

  const handleVinChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const val = e.target.value.toUpperCase();
    setVin(val);
    setVinError("");
    setVinSuccess(false);

    if (val.length === 17) {
      setIsValidatingVin(true);
      setTimeout(() => {
        setIsValidatingVin(false);
        setVinSuccess(true);
      }, 1200);
    }
  };

  if (!isOpen) return null;

  const nextStep = () => setStep((s) => Math.min(s + 1, 3));
  const prevStep = () => setStep((s) => Math.max(s - 1, 1));

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 backdrop-blur-sm p-4 animate-in fade-in duration-200">
      <div className="w-full max-w-2xl bg-surface-dark border border-border-color rounded-xl shadow-2xl flex flex-col max-h-[90vh] relative overflow-hidden">
        {/* Header */}
        <div className="px-8 py-6 border-b border-border-color bg-surface-darker/50">
          <div className="flex justify-between items-start mb-6">
            <div>
              <h2 className="text-xl font-bold text-white tracking-tight">
                {uiStrings?.add_vehicle_btn || "Add New Vehicle"}
              </h2>
              <p className="text-sm text-text-muted mt-1">
                {uiStrings?.add_vehicle_sub ||
                  "Initialize vehicle tracking in the global fleet."}
              </p>
            </div>
            <button
              onClick={onClose}
              className="text-text-muted hover:text-white transition-colors"
            >
              <span className="material-symbols-outlined">close</span>
            </button>
          </div>

          {/* Stepper */}
          <div className="relative flex items-center justify-between w-full max-w-md mx-auto">
            <div className="absolute left-0 top-[14px] w-full h-0.5 bg-border-color -z-10"></div>
            <div
              className={`absolute left-0 top-[14px] h-0.5 bg-primary -z-10 transition-all duration-300 ${step === 1 ? "w-0" : step === 2 ? "w-1/2" : "w-full"}`}
            ></div>

            {/* Step 1 */}
            <div className="flex flex-col items-center gap-2">
              <div
                className={`w-7 h-7 rounded-full flex items-center justify-center ring-4 ring-surface-dark transition-all ${step >= 1 ? "bg-primary text-white shadow-[0_0_10px_rgba(19,127,236,0.5)]" : "bg-surface-darker border border-border-color text-text-muted"}`}
              >
                {step > 1 ? (
                  <span className="material-symbols-outlined text-sm">
                    check
                  </span>
                ) : (
                  <span className="text-xs font-bold">1</span>
                )}
              </div>
              <span
                className={`text-[10px] uppercase font-bold tracking-wider ${step >= 1 ? "text-white" : "text-text-muted"}`}
              >
                {uiStrings?.step_basics || "Basics"}
              </span>
            </div>

            {/* Step 2 */}
            <div className="flex flex-col items-center gap-2">
              <div
                className={`w-7 h-7 rounded-full flex items-center justify-center ring-4 ring-surface-dark transition-all ${step >= 2 ? "bg-primary text-white shadow-[0_0_10px_rgba(19,127,236,0.5)]" : "bg-surface-darker border border-border-color text-text-muted"}`}
              >
                {step > 2 ? (
                  <span className="material-symbols-outlined text-sm">
                    check
                  </span>
                ) : (
                  <span className="text-xs font-bold">2</span>
                )}
              </div>
              <span
                className={`text-[10px] uppercase font-bold tracking-wider ${step >= 2 ? "text-white" : "text-text-muted"}`}
              >
                {uiStrings?.step_telematics || "Telematics"}
              </span>
            </div>

            {/* Step 3 */}
            <div className="flex flex-col items-center gap-2">
              <div
                className={`w-7 h-7 rounded-full flex items-center justify-center ring-4 ring-surface-dark transition-all ${step >= 3 ? "bg-primary text-white shadow-[0_0_10px_rgba(19,127,236,0.5)]" : "bg-surface-darker border border-border-color text-text-muted"}`}
              >
                <span className={`text-xs font-bold`}>3</span>
              </div>
              <span
                className={`text-[10px] uppercase font-bold tracking-wider ${step >= 3 ? "text-white" : "text-text-muted"}`}
              >
                {uiStrings?.step_driver || "Driver"}
              </span>
            </div>
          </div>
        </div>

        {/* Body */}
        <div className="p-8 bg-surface-dark flex-1 overflow-y-auto custom-scrollbar">
          {step === 1 && (
            <div className="space-y-6 animate-in slide-in-from-right-4 duration-300">
              <div className="space-y-2">
                <label className="block text-sm font-medium text-white">
                  {uiStrings?.vin_label ||
                    "Vehicle Identification Number (VIN)"}
                </label>
                <div className="relative">
                  <span className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none text-text-muted">
                    <span className="material-symbols-outlined text-[20px]">
                      fingerprint
                    </span>
                  </span>
                  <input
                    value={vin}
                    onChange={handleVinChange}
                    className={`block w-full rounded-lg bg-surface-darker border px-4 pl-10 py-2.5 text-white placeholder-text-muted/50 focus:ring-1 transition-all ${vinSuccess ? "border-emerald-500/50 focus:ring-emerald-500" : "border-border-color focus:ring-primary"}`}
                    placeholder={
                      uiStrings?.vin_placeholder || "e.g. 1HGCM82633A004352"
                    }
                    type="text"
                  />
                  {isValidatingVin && (
                    <div className="absolute right-3 top-2.5">
                      <span className="material-symbols-outlined animate-spin text-primary text-sm">
                        refresh
                      </span>
                    </div>
                  )}
                  {vinSuccess && (
                    <div className="absolute right-3 top-2.5">
                      <span className="material-symbols-outlined text-emerald-500 text-sm">
                        check_circle
                      </span>
                    </div>
                  )}
                </div>
                {vinSuccess && (
                  <div className="text-[11px] text-emerald-400 font-medium flex items-center gap-1.5 mt-1 bg-emerald-500/5 p-2 rounded border border-emerald-500/20">
                    <span className="material-symbols-outlined text-sm">
                      database
                    </span>
                    {uiStrings?.vin_validated || "NHTSA Validated"}: 2024
                    Freightliner Cascadia
                  </div>
                )}
              </div>
              <div className="grid grid-cols-2 gap-6">
                <div className="space-y-2">
                  <label className="block text-sm font-medium text-white">
                    {uiStrings?.license_label || "License Plate"}
                  </label>
                  <input
                    className="block w-full rounded-lg border-border-color bg-surface-darker px-4 py-2.5 text-white uppercase"
                    placeholder={uiStrings?.license_placeholder || "ABC-1234"}
                  />
                </div>
                <div className="space-y-2">
                  <label className="block text-sm font-medium text-white">
                    {uiStrings?.category_label || "Fleet Category"}
                  </label>
                  <select className="block w-full rounded-lg border-border-color bg-surface-darker px-4 py-2.5 text-white">
                    <option>Long Haul</option>
                    <option>Last Mile</option>
                    <option>Maintenance</option>
                  </select>
                </div>
              </div>
            </div>
          )}

          {step === 2 && (
            <div className="space-y-6 animate-in slide-in-from-right-4 duration-300">
              <div className="space-y-2">
                <label className="block text-sm font-medium text-white">
                  {uiStrings?.device_id_label || "Telematics Device ID"}
                </label>
                <input
                  className="block w-full rounded-lg border-border-color bg-surface-darker px-4 py-2.5 text-white"
                  placeholder={
                    uiStrings?.device_id_placeholder || "S/N: 9283-X92"
                  }
                />
              </div>
              <div className="space-y-2">
                <label className="block text-sm font-medium text-white">
                  {uiStrings?.install_date_label || "Installation Date"}
                </label>
                <div className="relative group">
                  <div className="absolute inset-y-0 left-3 flex items-center pointer-events-none">
                    <span className="material-symbols-outlined text-[14px] text-text-muted group-hover:text-primary transition-colors">
                      calendar_month
                    </span>
                  </div>
                  <input
                    type="date"
                    className="block w-full rounded-lg bg-surface-darker border border-border-color pl-10 pr-4 py-2.5 text-white [color-scheme:dark] shadow-sm outline-none transition-all hover:border-primary/30 focus:ring-1 focus:ring-primary/50 focus:border-primary"
                  />
                </div>
              </div>
            </div>
          )}

          {step === 3 && (
            <div className="space-y-6 animate-in slide-in-from-right-4 duration-300">
              <div className="space-y-2">
                <label className="block text-sm font-medium text-white">
                  {uiStrings?.driver_label || "Primary Driver"}
                </label>
                <select className="block w-full rounded-lg border-border-color bg-surface-darker px-4 py-2.5 text-white">
                  <option>{uiStrings?.unassigned || "Unassigned"}</option>
                  <option>Michael Johnson</option>
                  <option>Sarah Williams</option>
                </select>
              </div>
              <div className="space-y-2">
                <label className="block text-sm font-medium text-white">
                  {uiStrings?.org_label || "Service Organization"}
                </label>
                <select className="block w-full rounded-lg border-border-color bg-surface-darker px-4 py-2.5 text-white">
                  <option>Main HQ Depot</option>
                  <option>West Coast Branch</option>
                </select>
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="px-8 py-5 border-t border-border-color bg-surface-darker flex justify-between items-center">
          <button
            onClick={step === 1 ? onClose : prevStep}
            className="px-4 py-2.5 text-sm font-medium text-text-muted hover:text-white transition-colors"
          >
            {step === 1
              ? uiStrings?.cancel_btn || "Cancel"
              : uiStrings?.back_btn || "Back"}
          </button>
          <button
            onClick={step === 3 ? onClose : nextStep}
            className="px-6 py-2.5 text-sm font-medium text-white bg-primary rounded-lg hover:bg-primary/90 shadow-[0_0_15px_rgba(19,127,236,0.3)] flex items-center gap-2"
          >
            {step === 3
              ? uiStrings?.finish_btn || "Finish & Add"
              : uiStrings?.next_btn || "Next Step"}
            <span className="material-symbols-outlined text-sm">
              arrow_forward
            </span>
          </button>
        </div>
      </div>
    </div>
  );
};
