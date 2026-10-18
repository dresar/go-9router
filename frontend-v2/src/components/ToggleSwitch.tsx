import React from "react"

interface ToggleSwitchProps {
  checked: boolean
  onChange: (checked: boolean) => void
  disabled?: boolean
  label?: string
  size?: "sm" | "md"
}

export const ToggleSwitch: React.FC<ToggleSwitchProps> = ({
  checked,
  onChange,
  disabled = false,
  label = "Toggle status",
  size = "md"
}) => {
  const handleClick = (e: React.MouseEvent) => {
    e.stopPropagation()
    if (!disabled) {
      onChange(!checked)
    }
  }

  const isSmall = size === "sm"
  const widthClass = isSmall ? "w-8 h-4" : "w-9 h-5"
  const thumbClass = isSmall ? "w-3 h-3" : "w-4 h-4"
  const translateClass = isSmall
    ? checked ? "translate-x-4" : "translate-x-0.5"
    : checked ? "translate-x-4" : "translate-x-0.5"

  return (
    <button
      type="button"
      role="switch"
      aria-checked={checked}
      aria-label={label}
      disabled={disabled}
      onClick={handleClick}
      className={`relative inline-flex ${widthClass} shrink-0 cursor-pointer items-center rounded-full border border-transparent transition-colors duration-200 ease-in-out focus:outline-none focus-visible:ring-2 focus-visible:ring-emerald-500 disabled:cursor-not-allowed disabled:opacity-50 ${
        checked ? "bg-emerald-600 border-emerald-500" : "bg-[#27272a] border-[#3f3f46]"
      }`}
    >
      <span
        className={`pointer-events-none inline-block ${thumbClass} transform rounded-full bg-white shadow-xs transition duration-200 ease-in-out ${translateClass}`}
      />
    </button>
  )
}
