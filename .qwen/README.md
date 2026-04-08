# MCP Configuration for FlyFleet

This directory contains the MCP (Model Context Protocol) server configuration for the FlyFleet project.

## 📁 Files

- `mcp.json` - MCP server configuration
- `../.env.example` - Environment variable template
- `../.env` - Your local environment (create from template)
- `../scripts/setup-mcp.sh` - Setup helper script

## 🔧 Stitch MCP Server

The FlyFleet project uses **Google Stitch MCP** for enhanced capabilities:

- **Transport**: HTTP
- **URL**: https://stitch.googleapis.com/mcp
- **Authentication**: Google Cloud API Key

## 🚀 Quick Setup

### Option 1: Run Setup Script (Recommended)

```bash
./scripts/setup-mcp.sh
```

Then edit the `.env` file with your API key.

### Option 2: Manual Setup

1. **Create `.env` file:**
   ```bash
   cp .env.example .env
   ```

2. **Edit `.env` and add your API key:**
   ```
   GOOGLE_API_KEY=your-actual-api-key-here
   ```

3. **Source the environment:**
   ```bash
   export $(cat .env | xargs)
   ```

### Option 3: System-Wide Environment Variable

Add to your shell profile (`~/.bashrc` or `~/.zshrc`):

```bash
export GOOGLE_API_KEY="your-actual-api-key"
```

Then reload your shell:

```bash
source ~/.bashrc  # or source ~/.zshrc
```

## 🔑 Getting Your API Key

1. Go to [Google Cloud Console](https://console.cloud.google.com/apis/credentials)
2. Create a new API key or use an existing one
3. Copy the API key
4. Add it to your `.env` file or environment variable

## ✅ Verification

After configuration, the Stitch MCP server will be available when you start Claude Code.

To verify, check if the MCP server is listed in your Claude Code MCP servers.

## 📝 Configuration Format

The `mcp.json` uses this format:

```json
{
  "mcpServers": {
    "stitch": {
      "transport": {
        "type": "http",
        "url": "https://stitch.googleapis.com/mcp"
      },
      "headers": {
        "X-Goog-Api-Key": "${GOOGLE_API_KEY}"
      }
    }
  }
}
```

## 🔒 Security Best Practices

- ✅ **DO** use environment variables for API keys
- ✅ **DO** add `.env` to `.gitignore` (already done)
- ❌ **DON'T** commit API keys to version control
- ❌ **DON'T** share API keys in chat or public forums
- 🔄 **DO** rotate keys periodically

## 🛠️ Troubleshooting

### MCP Server Not Loading

1. Check if `.env` file exists and contains valid API key
2. Verify the API key is correct (no extra spaces)
3. Restart Claude Code
4. Check Claude Code logs for errors

### API Key Errors

1. Verify the API key is valid in Google Cloud Console
2. Check if the API is enabled for your project
3. Ensure the key has necessary permissions
4. Try creating a new API key

## 📚 Related Documentation

- [Stitch MCP Documentation](https://stitch.google.dev/docs)
- [Google Cloud API Keys](https://cloud.google.com/docs/authentication/api-keys)
- [MCP Specification](https://modelcontextprotocol.io/)

---

**Created:** March 31, 2026  
**Maintained By:** Engineering Team
