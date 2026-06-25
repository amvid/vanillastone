# Dev image: Vite dev server with HMR. Source mounted at runtime (see compose).
FROM node:24-alpine AS dev

WORKDIR /app

# corepack ships with node; activates the pnpm version pinned in package.json.
RUN corepack enable

# Cache deps first. Lockfile optional on first run — glob avoids COPY failure.
COPY package.json pnpm-lock.yaml* ./
RUN pnpm install

COPY . .

EXPOSE 5173

CMD ["pnpm", "dev"]
