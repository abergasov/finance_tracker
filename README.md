local run in docker
```shell
make run
```

Google OAuth login stays disabled until all of these backend settings in `configs/app_conf.yml` are real values:
- `auth.ui_base_url` must be the absolute URL where the UI is served, for example `http://localhost:3000`
- `auth.google.client_id`
- `auth.google.client_secret`
- `auth.google.redirect_url`
- `auth.token.signing_key`

The shipped placeholder values (`your-google-client-id`, `your-google-client-secret`, `replace-with-a-long-random-secret`) intentionally keep auth disabled.

### Running the UI against the backend

The UI can still use the Vite `/api` proxy in local dev, but it no longer depends on that path.

- when the UI and backend share the same origin, no extra UI config is required
- when the UI runs separately, set `ui/.env` with `PUBLIC_API_BASE_URL=http://localhost:8000`
- keep backend `auth.ui_base_url` pointed at the UI origin so Google callback completion can redirect back to `/auth/callback`, and so backend CORS allows browser auth requests from that origin

Example local setup:

```shell
# backend
go run ./cmd -config ./configs/app_conf.yml

# ui
cd ui
cp .env.example .env
npm install
npm run dev
```
