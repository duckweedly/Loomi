# Quickstart: M16 Work Mode Foundation

1. Run web tests:

   ```sh
   bun test --cwd web
   ```

2. Build web:

   ```sh
   bun run --cwd web build
   ```

3. Build docs:

   ```sh
   bun run --cwd docs-site build
   ```

4. Run diff whitespace check:

   ```sh
   git diff --check
   ```

5. Browser smoke:

   ```sh
   bun run --cwd web dev -- --port 5180 --strictPort
   ```

   Open `http://127.0.0.1:5180`, select Work mode, verify the Work Plan View shows goal, steps, progress, artifacts, and recent events. Switch to Chat mode and verify normal chat remains without the Work Plan View.
