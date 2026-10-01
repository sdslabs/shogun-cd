# Shogun-CD UI

The web interface for Shogun-CD. It is a React and TypeScript application built with Vite, Tailwind CSS, shadcn-style components, and TanStack Query.

## Development

```sh
npm install
npm run dev
```

Vite listens on `http://localhost:5173` and proxies requests under `/api` to the Go API at `http://localhost:7007`.

Use environment variables when those defaults do not fit:

```sh
cp .env.example .env.local
```

- `VITE_API_BASE_URL` is the API prefix used in the browser. It defaults to `/api`.
- `SHOGUN_API_PROXY_TARGET` is the Go API used by the development proxy.

## Checks

```sh
npm run typecheck
npm run lint
npm test
npm run build
```

Pipeline and target definitions are intentionally read-only. Git remains their source of truth.
