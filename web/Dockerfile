# Dev image: Vite dev server with HMR. Source mounted at runtime (see compose).
FROM denoland/deno:2.9.1 AS dev

WORKDIR /app

# Cache deps first. Deno reads npm deps from package.json; nodeModulesDir=auto
# (deno.json) materializes them into ./node_modules.
COPY package.json deno.json ./
RUN deno install

COPY . .

EXPOSE 5173

# --host binds 0.0.0.0 so the port is reachable from outside the container.
CMD ["deno", "task", "dev", "--host"]
