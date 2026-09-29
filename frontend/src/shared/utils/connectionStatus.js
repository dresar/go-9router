export function getStatusVariant(isActive, effectiveStatus) {
  if (isActive === false) return "default";
  if (effectiveStatus === "active" || effectiveStatus === "success" || effectiveStatus === "available") return "success";
  if (effectiveStatus === "error" || effectiveStatus === "expired" || effectiveStatus === "unavailable") return "error";
  return "default";
}
