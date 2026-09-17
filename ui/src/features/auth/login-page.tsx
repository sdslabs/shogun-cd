import { zodResolver } from "@hookform/resolvers/zod"
import { ArrowRight, Eye, EyeOff } from "lucide-react"
import * as React from "react"
import { useForm } from "react-hook-form"
import { Link, useLocation, useNavigate } from "react-router-dom"
import { z } from "zod"

import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { useAuth } from "@/features/auth/auth-provider"

const schema = z.object({
  email: z.email("Enter a valid email address"),
  password: z.string().min(1, "Enter your password"),
})

type FormValues = z.infer<typeof schema>

export function LoginPage() {
  const { login } = useAuth()
  const navigate = useNavigate()
  const location = useLocation()
  const [showPassword, setShowPassword] = React.useState(false)
  const [serverError, setServerError] = React.useState<string | null>(null)
  const { register, handleSubmit, formState: { errors, isSubmitting } } = useForm<FormValues>({ resolver: zodResolver(schema), defaultValues: { email: "", password: "" } })

  const onSubmit = handleSubmit(async (values) => {
    setServerError(null)
    try {
      await login(values)
      const destination = (location.state as { from?: { pathname?: string } } | null)?.from?.pathname || "/pipelines"
      navigate(destination, { replace: true })
    } catch (error) {
      setServerError(error instanceof Error ? error.message : "Unable to sign in")
    }
  })

  return (
    <div>
      <p className="text-[10px] font-semibold tracking-[0.2em] text-primary uppercase">Welcome back</p>
      <h2 className="mt-3 text-3xl font-semibold tracking-[-0.045em]">Sign in to Shogun</h2>
      <p className="mt-2 text-sm leading-6 text-muted-foreground">Return to your delivery workspace.</p>

      <form className="mt-8 space-y-5" onSubmit={onSubmit} noValidate>
        {serverError ? <div role="alert" className="rounded-lg border border-destructive/20 bg-destructive/7 px-3.5 py-3 text-sm text-destructive">{serverError}</div> : null}
        <Field label="Email" error={errors.email?.message}><Input autoComplete="email" autoFocus placeholder="you@company.com" {...register("email")} /></Field>
        <Field label="Password" error={errors.password?.message}>
          <div className="relative"><Input type={showPassword ? "text" : "password"} autoComplete="current-password" className="pr-10" {...register("password")} /><button type="button" className="absolute top-1/2 right-3 -translate-y-1/2 text-muted-foreground hover:text-foreground" onClick={() => setShowPassword((value) => !value)} aria-label={showPassword ? "Hide password" : "Show password"}>{showPassword ? <EyeOff className="size-4" /> : <Eye className="size-4" />}</button></div>
        </Field>
        <Button type="submit" size="lg" className="w-full" disabled={isSubmitting}>{isSubmitting ? "Signing in…" : <>Sign in <ArrowRight /></>}</Button>
      </form>
      <p className="mt-7 text-center text-sm text-muted-foreground">New to Shogun? <Link to="/register" className="font-medium text-foreground underline-offset-4 hover:underline">Create an account</Link></p>
    </div>
  )
}

function Field({ label, error, children }: { label: string; error?: string; children: React.ReactNode }) {
  return <label className="block"><span className="mb-2 block text-xs font-medium">{label}</span>{children}{error ? <span className="mt-1.5 block text-xs text-destructive">{error}</span> : null}</label>
}
