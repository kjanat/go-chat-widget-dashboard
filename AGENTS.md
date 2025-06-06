# Repository Guidelines

This project is structured with Go handlers. HTML components live in `internal/templates` as `.templ` files generated with templ.

## Development

- Install dependencies with `make install`.
- Initialize a local database with `make init-db`.
- Run the application with `make run`.
- To start a development server with hot reload, use `make dev` (requires `air`).

## Testing and Linting

- Always run `make test` before committing to ensure all packages compile and tests pass.
- Format and lint code with `make fmt` and `make lint`.

## Templates and Templ

- HTML templates live in `web/templates` and are loaded with `html/template` (standard Go templates).
- New `.templ` files belong in `internal/templates` and use the templ DSL.
- After creating or editing a `.templ` file, run:
  ```bash
  templ generate ./internal/templates/*.templ
  ```
  This produces Go source files which must be committed.
- Use the generated components in handlers instead of `ExecuteTemplate`.

## CSS with Tailwind

- Source CSS is in `web/static/css/tailwind.css` which should contain `@tailwind` directives.
- Compile it to `web/static/css/admin.css` using:

  ```bash
  npx tailwindcss -i web/static/css/tailwind.css -o web/static/css/admin.css --watch
  ```

