import { useState } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"


import { deleteStore, getStores } from "@/api/stores"
import { Button } from "@/components/ui/button"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from "@/components/ui/alert-dialog"
import { CreateStoreModal } from "@/components/domain/create-store-modal"
import { StatusBadge } from "@/components/domain/status-badge"

const storesQueryKey = ["stores"] as const

export default function Dashboard() {
  const queryClient = useQueryClient()
  const [storeToDelete, setStoreToDelete] = useState<string | null>(null)
  const [deletingStores, setDeletingStores] = useState<Set<string>>(new Set())

  const storesQuery = useQuery({
    queryKey: storesQueryKey,
    queryFn: getStores,
    refetchInterval: 5000,
  })

  const deleteMutation = useMutation({
    mutationFn: (name: string) => deleteStore(name),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: storesQueryKey })
    },
    onError: (_err, name) => {
      // If deletion fails, remove from deleting set
      setDeletingStores((prev) => {
        const next = new Set(prev)
        next.delete(name)
        return next
      })
    },
    onSettled: () => {
      setStoreToDelete(null)
    },
  })

  const handleDeleteConfirm = () => {
    if (storeToDelete) {
      // Optimistically mark as deleting
      setDeletingStores((prev) => new Set(prev).add(storeToDelete))
      deleteMutation.mutate(storeToDelete)
    }
  }

  const stores = storesQuery.data ?? []

  return (
    <div className="mx-auto max-w-5xl p-6 space-y-6">
      <div className="flex items-center justify-between gap-3">
        <div className="space-y-1">
          <h1 className="text-2xl font-semibold">Stores</h1>
          <p className="text-sm text-muted-foreground">
            Manage Kubernetes Store resources.
          </p>
        </div>
        <CreateStoreModal />
      </div>

      {storesQuery.isError && (
        <div className="rounded-md border p-3 text-sm text-destructive">
          {storesQuery.error.message}
        </div>
      )}

      <div className="rounded-md border">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead>Name</TableHead>
              <TableHead>Plan</TableHead>
              <TableHead>Status</TableHead>
              <TableHead>URL</TableHead>
              <TableHead className="text-right">Actions</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {storesQuery.isLoading ? (
              <TableRow>
                <TableCell colSpan={5} className="text-muted-foreground">
                  Loading...
                </TableCell>
              </TableRow>
            ) : stores.length === 0 ? (
              <TableRow>
                <TableCell colSpan={5} className="text-muted-foreground">
                  No stores found.
                </TableCell>
              </TableRow>
            ) : (
              stores.map((s) => (
                <TableRow key={`${s.namespace}/${s.name}`}>
                  <TableCell className="font-medium">{s.name}</TableCell>
                  <TableCell>{s.plan}</TableCell>
                  <TableCell>
                    <StatusBadge
                      status={deletingStores.has(s.name) ? "Deleting" : s.status}
                    />
                  </TableCell>
                  <TableCell>
                    {s.url ? (
                      <a
                        href={s.url}
                        target="_blank"
                        rel="noreferrer"
                        className="text-primary underline underline-offset-4"
                      >
                        Open
                      </a>
                    ) : (
                      <span className="text-muted-foreground">—</span>
                    )}
                  </TableCell>
                  <TableCell className="text-right">
                    <Button
                      variant="destructive"
                      size="sm"
                      disabled={deletingStores.has(s.name)}
                      onClick={() => setStoreToDelete(s.name)}
                    >
                      Delete
                    </Button>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      <AlertDialog open={!!storeToDelete} onOpenChange={(open) => !open && setStoreToDelete(null)}>
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>Delete Store</AlertDialogTitle>
            <AlertDialogDescription>
              Are you sure you want to delete <strong>{storeToDelete}</strong>? This action cannot be undone.
            </AlertDialogDescription>
          </AlertDialogHeader>
          <AlertDialogFooter>
            <AlertDialogCancel disabled={deleteMutation.isPending}>Cancel</AlertDialogCancel>
            <AlertDialogAction
              onClick={handleDeleteConfirm}
              disabled={deleteMutation.isPending}
              className="bg-destructive text-destructive-foreground hover:bg-destructive/90"
            >
              {deleteMutation.isPending ? "Deleting..." : "Delete"}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  )
}
