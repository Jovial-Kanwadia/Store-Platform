import { Badge } from "@/components/ui/badge"
import { cn } from "@/lib/utils"

export function StatusBadge({ status }: { status: unknown }) {
  // Safe handling if status is an object (old backend response) or undefined
  const statusStr = typeof status === 'string' ? status : String(status || "")
  const normalized = statusStr.trim()

  if (normalized === "Ready") {
    return (
      <Badge className={cn("bg-emerald-600 text-white hover:bg-emerald-600")}>
        Ready
      </Badge>
    )
  }

  if (normalized === "Provisioning") {
    return (
      <Badge
        className={cn(
          "bg-amber-500 text-black hover:bg-amber-500 animate-pulse"
        )}
      >
        Provisioning
      </Badge>
    )
  }

  if (normalized === "Failed") {
    return (
      <Badge variant="destructive" className={cn("hover:bg-destructive")}>
        Failed
      </Badge>
    )
  }

  if (normalized === "Deleting") {
    return (
      <Badge className={cn("bg-red-600 text-white hover:bg-red-600 animate-pulse")}>
        Deleting
      </Badge>
    )
  }

  return <Badge variant="secondary">{normalized || "Unknown"}</Badge>
}

