=== Rocky Ads Web Site ===

Development
-----------

**Prerequisites:** Node.js (includes npm). Install from https://nodejs.org or e.g. `brew install node` on macOS.

**CSS (Tailwind)**

The app uses Tailwind CSS v4 (listed in `package.json`; no need to install it separately). Install deps once, then build the stylesheet before or while running the server:

1. `npm install`
2. One-off build: `npm run build-css-dev`
   - Or watch mode (rebuild on changes): `npm run build-css`
   - Or production (minified): `npm run build-css-prod`

Input: `input.css` → output: `static/css/output.css`.
