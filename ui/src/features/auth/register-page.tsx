import { zodResolver } from "@hookform/resolvers/zod"
import { ArrowRight, CheckCircle2 } from "lucide-react"
import * as React from "react"
import { useForm } from "react-hook-form"
import { Link } from "react-router-dom"
import { z } from "zod"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { api } from "@/lib/api/endpoints"

const schema = z.object({
  email: z.email("Enter a valid email address"),
  password: z.string().min(8, "Use at least 8 characters"),
  confirmPassword: z.string(),
}).refine((values) => values.password === values.confirmPassword, { message: "Passwords do not match", path: ["confirmPassword"] })

type FormValues = z.infer<typeof schema>

export function RegisterPage() {
  const [complete, setComplete] = React.useState(false)
  const [serverError, setServerError] = React.useState<string | null>(null)
  const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm<FormValues>({ resolver: zodResolver(schema), defaultValues: { email: "", password: "", confirmPassword: "" } })
  const onSubmit = handleSubmit(async ({ email, password }) => {
    setServerError(null)
    try { await api.register({ email, password }); setComplete(true) } catch (error) { setServerError(error instanceof Error ? error.message : "Unable to create account") }
  })

  if (complete) return (
    <div className="text-center">
      <span className="mx-auto flex size-12 items-center justify-center rounded-full bg-success/10 text-success"><CheckCircle2 className="size-6" /></span>
      <h2 className="mt-5 text-2xl font-semibold tracking-[-0.035em]">Account created</h2>
      <p className="mt-2 text-sm leading-6 text-muted-foreground">Your account is ready. Sign in to open the delivery workspace.</p>
      <Button asChild className="mt-7"><Link to="/login">Continue to sign in <ArrowRight /></Link></Button>
    </div>
  )

  return (
    <div>
      <p className="text-[10px] font-semibold tracking-[0.2em] text-primary uppercase">Join your team</p>
      <h2 className="mt-3 text-3xl font-semibold tracking-[-0.045em]">Create an account</h2>
      <p className="mt-2 text-sm leading-6 text-muted-foreground">Set up your access to this Shogun instance.</p>
      <form className="mt-8 space-y-4" onSubmit={onSubmit} noValidate>
        {serverError ? <div role="alert" className="rounded-lg border border-destructive/20 bg-destructive/7 px-3.5 py-3 text-sm text-destructive">{serverError}</div> : null}
        <Field label="Email" error={errors.email?.message}><Input autoComplete="email" autoFocus placeholder="you@company.com" {...register("email")} /></Field>
        <Field label="Password" error={errors.password?.message}><Input type="password" autoComplete="new-password" {...register("password")} /></Field>
        <Field label="Confirm password" error={errors.confirmPassword?.message}><Input type="password" autoComplete="new-password" {...register("confirmPassword")} /></Field>
        <Button type="submit" size="lg" className="mt-2 w-full" disabled={isSubmitting}>{isSubmitting ? "Creating account…" : <>Create account <ArrowRight /></>}</Button>
      </form>
      <p className="mt-7 text-center text-sm text-muted-foreground">Already have an account? <Link to="/login" className="font-medium text-foreground underline-offset-4 hover:underline">Sign in</Link></p>
    </div>
  )
}

function Field({ label, error, children }: { label: string; error?: string; children: React.ReactNode }) {
  return <label className="block"><span className="mb-2 block text-xs font-medium">{label}</span>{children}{error ? <span className="mt-1.5 block text-xs text-destructive">{error}</span> : null}</label>
}
