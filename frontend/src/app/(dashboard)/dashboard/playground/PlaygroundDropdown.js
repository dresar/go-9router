"use client";

import { useState, useRef, useEffect } from "react";
import PropTypes from "prop-types";
import ProviderIcon from "@/shared/components/ProviderIcon";
import { cn } from "@/shared/utils/cn";

export default function PlaygroundDropdown({
  label,
  value,
  onChange,
  options = [],
  placeholder = "Select...",
  searchable = true,
  searchPlaceholder = "Search...",
  disabled = false,
  className = "",
  buttonClassName = "",
}) {
  const [isOpen, setIsOpen] = useState(false);
  const [search, setSearch] = useState("");
  const containerRef = useRef(null);
  const searchInputRef = useRef(null);

  // Close on outside click
  useEffect(() => {
    const handleClickOutside = (e) => {
      if (containerRef.current && !containerRef.current.contains(e.target)) {
        setIsOpen(false);
      }
    };
    if (isOpen) {
      document.addEventListener("mousedown", handleClickOutside);
    }
    return () => {
      document.removeEventListener("mousedown", handleClickOutside);
    };
  }, [isOpen]);

  // Focus search input on open
  useEffect(() => {
    if (isOpen && searchable && searchInputRef.current) {
      searchInputRef.current.focus();
    }
    if (!isOpen) {
      setSearch("");
    }
  }, [isOpen, searchable]);

  // Find selected item
  const selectedOption = options.find((opt) => opt.id === value || opt.value === value);

  // Filter options by search
  const filteredOptions = options.filter((opt) => {
    if (!search.trim()) return true;
    const term = search.toLowerCase();
    const nameMatch = opt.name?.toLowerCase().includes(term);
    const idMatch = opt.id?.toLowerCase().includes(term);
    const subtitleMatch = opt.subtitle?.toLowerCase().includes(term);
    return nameMatch || idMatch || subtitleMatch;
  });

  const handleSelect = (opt) => {
    onChange(opt.id || opt.value);
    setIsOpen(false);
  };

  return (
    <div className={cn("relative flex flex-col gap-1.5", className)} ref={containerRef}>
      {label && (
        <label className="text-xs font-medium text-text-muted flex items-center justify-between">
          <span>{label}</span>
        </label>
      )}

      {/* Trigger Button */}
      <button
        type="button"
        disabled={disabled}
        onClick={() => setIsOpen((prev) => !prev)}
        className={cn(
          "w-full h-9 px-3 flex items-center justify-between gap-2",
          "bg-surface-2 hover:bg-surface-3 border border-border-subtle hover:border-brand-500/40",
          "rounded-[10px] text-xs text-text-main font-medium transition-all text-left",
          "focus:outline-none focus:ring-1 focus:ring-brand-500/40 focus:border-brand-500/60",
          disabled && "opacity-50 cursor-not-allowed",
          isOpen && "border-brand-500/50 ring-1 ring-brand-500/30",
          buttonClassName
        )}
      >
        <div className="flex items-center gap-2 min-w-0">
          {selectedOption ? (
            <>
              {selectedOption.isCombo ? (
                <span className="material-symbols-outlined text-[18px] text-amber-400 shrink-0">
                  bolt
                </span>
              ) : selectedOption.providerId || selectedOption.provider ? (
                <ProviderIcon
                  providerId={selectedOption.providerId || selectedOption.provider}
                  size={18}
                  className="rounded shrink-0 object-contain"
                />
              ) : selectedOption.icon ? (
                <span className="material-symbols-outlined text-[18px] text-text-muted shrink-0">
                  {selectedOption.icon}
                </span>
              ) : null}

              <span className="truncate">{selectedOption.name || selectedOption.id}</span>

              {selectedOption.badge && (
                <span className="shrink-0 text-[10px] px-1.5 py-0.2 font-mono font-medium rounded bg-brand-500/10 text-brand-500 border border-brand-500/20">
                  {selectedOption.badge}
                </span>
              )}
            </>
          ) : (
            <span className="text-text-muted truncate">{placeholder}</span>
          )}
        </div>

        <span
          className={cn(
            "material-symbols-outlined text-[18px] text-text-muted transition-transform shrink-0",
            isOpen && "rotate-180 text-brand-500"
          )}
        >
          expand_more
        </span>
      </button>

      {/* Floating Dropdown Menu */}
      {isOpen && (
        <div
          className={cn(
            "absolute left-0 right-0 top-full mt-1.5 z-50",
            "bg-surface border border-border rounded-[12px] shadow-2xl",
            "py-1 min-w-[220px] max-h-72 flex flex-col",
            "animate-in fade-in zoom-in-95 duration-100"
          )}
        >
          {/* Optional Search */}
          {searchable && options.length > 5 && (
            <div className="p-1.5 border-b border-border-subtle">
              <div className="relative">
                <input
                  ref={searchInputRef}
                  type="text"
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  placeholder={searchPlaceholder}
                  className="w-full bg-surface-2 border border-border-subtle rounded-md pl-7 pr-2 py-1 text-xs text-text-main placeholder-text-muted focus:outline-none focus:border-brand-500/50"
                />
                <span className="material-symbols-outlined text-[14px] text-text-muted absolute left-2 top-1.5">
                  search
                </span>
              </div>
            </div>
          )}

          {/* Options List */}
          <div className="overflow-y-auto custom-scrollbar flex-1 py-1">
            {filteredOptions.length === 0 ? (
              <div className="px-3 py-3 text-center text-xs text-text-muted">
                No data
              </div>
            ) : (
              filteredOptions.map((opt) => {
                const isSelected = (opt.id || opt.value) === value;
                return (
                  <button
                    key={opt.id || opt.value}
                    type="button"
                    onClick={() => handleSelect(opt)}
                    className={cn(
                      "w-full px-3 py-2 flex items-center justify-between gap-2 text-xs text-left transition-colors",
                      isSelected
                        ? "bg-brand-500/10 text-brand-500 font-medium"
                        : "text-text-main hover:bg-surface-2"
                    )}
                  >
                    <div className="flex items-center gap-2 min-w-0">
                      {opt.isCombo ? (
                        <span className="material-symbols-outlined text-[16px] text-amber-400 shrink-0">
                          bolt
                        </span>
                      ) : opt.providerId || opt.provider ? (
                        <ProviderIcon
                          providerId={opt.providerId || opt.provider}
                          size={18}
                          className="rounded shrink-0 object-contain"
                        />
                      ) : opt.icon ? (
                        <span className="material-symbols-outlined text-[16px] text-text-muted shrink-0">
                          {opt.icon}
                        </span>
                      ) : null}

                      <div className="min-w-0 flex flex-col">
                        <span className="truncate">{opt.name || opt.id}</span>
                        {opt.subtitle && (
                          <span className="text-[10px] text-text-muted truncate">
                            {opt.subtitle}
                          </span>
                        )}
                      </div>
                    </div>

                    <div className="flex items-center gap-1 shrink-0">
                      {opt.badge && (
                        <span className="text-[10px] px-1.5 py-0.2 rounded font-mono bg-surface-3 text-text-muted border border-border-subtle">
                          {opt.badge}
                        </span>
                      )}
                      {isSelected && (
                        <span className="material-symbols-outlined text-[16px] text-brand-500">
                          check
                        </span>
                      )}
                    </div>
                  </button>
                );
              })
            )}
          </div>
        </div>
      )}
    </div>
  );
}

PlaygroundDropdown.propTypes = {
  label: PropTypes.string,
  value: PropTypes.string,
  onChange: PropTypes.func.isRequired,
  options: PropTypes.array,
  placeholder: PropTypes.string,
  searchable: PropTypes.bool,
  searchPlaceholder: PropTypes.string,
  disabled: PropTypes.bool,
  className: PropTypes.string,
  buttonClassName: PropTypes.string,
};
