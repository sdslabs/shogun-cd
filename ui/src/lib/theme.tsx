import * as React from "react"

type Theme = "light" | "dark" | "system"

interface ThemeContextValue {
  theme: Theme
  setTheme: (theme: Theme) => void
  resolvedTheme: "light" | "dark"
}

const ThemeContext = React.createContext<ThemeContextValue | null>(null)
const THEME_KEY = "shogun.theme"

function resolveTheme(theme: Theme) {
  if (theme !== "system") return theme
  return window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light"
}

export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setTheme] = React.useState<Theme>(() => {
    const stored = window.localStorage.getItem(THEME_KEY)
    return stored === "light" || stored === "dark" || stored === "system" ? stored : "system"
  })
  const [resolvedTheme, setResolvedTheme] = React.useState<"light" | "dark">(() => resolveTheme(theme))

  React.useEffect(() => {
    const media = window.matchMedia("(prefers-color-scheme: dark)")
    const update = () => {
      const resolved = resolveTheme(theme)
      setResolvedTheme(resolved)
      document.documentElement.classList.toggle("dark", resolved === "dark")
      document.documentElement.style.colorScheme = resolved
    }
    update()
    media.addEventListener("change", update)
    return () => media.removeEventListener("change", update)
  }, [theme])

  const updateTheme = React.useCallback((nextTheme: Theme) => {
    window.localStorage.setItem(THEME_KEY, nextTheme)
    setTheme(nextTheme)
  }, [])

  return (
    <ThemeContext.Provider value={{ theme, setTheme: updateTheme, resolvedTheme }}>
      {children}
    </ThemeContext.Provider>
  )
}

export function useTheme() {
  const context = React.useContext(ThemeContext)
  if (!context) throw new Error("useTheme must be used inside ThemeProvider")
  return context
}
