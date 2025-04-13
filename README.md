# GoShareX

ShareX uploader service built with TinyGo and Cloudflare Workers

## Setup

1. Clone the repository
2. Copy `.dev.vars.example` to `.dev.vars` and set the environment variables
3. Copy `wrangler.jsonc.example` to `wrangler.jsonc` change the required fields
4. Install dependencies: `go mod download`
5. Run locally: `npm run start`
6. Create the secret `wrangler secret put SHAREX_AUTH_KEY`
6. Deploy: `npm run deploy`

## Usage

1. Visit `/config?authKey=your_auth_key` to download the ShareX configuration
2. Import the configuration into ShareX
3. Start uploading images
