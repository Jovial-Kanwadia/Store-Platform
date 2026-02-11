import { useMemo, useState } from "react"
import { useForm, Controller } from "react-hook-form"
import { z } from "zod"
import { useMutation, useQueryClient } from "@tanstack/react-query"

import { createStore } from "@/api/stores"
import type { CreateStoreRequest } from "@/types"
import { Button } from "@/components/ui/button"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"

type FormValues = {
  name: string
  plan: "small" | "medium"
  engine: "woo"
}

const schema = z.object({
  name: z
    .string()
    .min(1, "Name is required")
    .regex(/^[a-z0-9]+$/, "Lowercase alphanumeric only"),
  plan: z.enum(["small", "medium"]),
  engine: z.enum(["woo"]),
})

export function CreateStoreModal() {
  const [open, setOpen] = useState(false)
  const queryClient = useQueryClient()

  const defaultValues = useMemo<FormValues>(
    () => ({ name: "", plan: "small", engine: "woo" }),
    []
  )

  const {
    register,
    handleSubmit,
    control,
    formState: { errors },
    setError,
    reset,
  } = useForm<FormValues>({ defaultValues })

  const mutation = useMutation({
    mutationFn: (payload: CreateStoreRequest) => createStore(payload),
    onSuccess: async () => {
      setOpen(false)
      reset(defaultValues)
      await queryClient.invalidateQueries({ queryKey: ["stores"] })
    },
  })

  const onSubmit = handleSubmit((values) => {
    const parsed = schema.safeParse(values)
    if (!parsed.success) {
      for (const issue of parsed.error.issues) {
        const field = issue.path[0]
        if (field === "name" || field === "plan" || field === "engine") {
          setError(field, { type: "validate", message: issue.message })
        }
      }
      return
    }

    mutation.mutate({
      name: parsed.data.name,
      plan: parsed.data.plan,
      engine: parsed.data.engine,
    })
  })

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button>Create store</Button>
      </DialogTrigger>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Create store</DialogTitle>
          <DialogDescription>
            Provision a new store in your Kubernetes cluster.
          </DialogDescription>
        </DialogHeader>

        <form className="grid gap-4" onSubmit={onSubmit}>
          <div className="grid gap-2">
            <Label htmlFor="name">Name</Label>
            <Input
              id="name"
              placeholder="my-store"
              autoComplete="off"
              {...register("name")}
            />
            {errors.name?.message && (
              <p className="text-sm text-destructive">{errors.name.message}</p>
            )}
          </div>

          <div className="grid gap-2">
            <Label>Plan</Label>
            <Controller
              control={control}
              name="plan"
              render={({ field }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Select plan" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="small">small</SelectItem>
                    <SelectItem value="medium">medium</SelectItem>
                  </SelectContent>
                </Select>
              )}
            />
            {errors.plan?.message && (
              <p className="text-sm text-destructive">{errors.plan.message}</p>
            )}
          </div>

          <div className="grid gap-2">
            <Label>Engine</Label>
            <Controller
              control={control}
              name="engine"
              render={({ field }) => (
                <Select value={field.value} onValueChange={field.onChange}>
                  <SelectTrigger className="w-full">
                    <SelectValue placeholder="Select engine" />
                  </SelectTrigger>
                  <SelectContent>
                    <SelectItem value="woo">woo</SelectItem>
                  </SelectContent>
                </Select>
              )}
            />
            {errors.engine?.message && (
              <p className="text-sm text-destructive">
                {errors.engine.message}
              </p>
            )}
          </div>

          <DialogFooter>
            <Button type="submit" disabled={mutation.isPending}>
              {mutation.isPending ? "Creating..." : "Create"}
            </Button>
          </DialogFooter>

          {mutation.error?.message && (
            <p className="text-sm text-destructive">{mutation.error.message}</p>
          )}
        </form>
      </DialogContent>
    </Dialog>
  )
}

