# Qoder MCP Server Orchestrator

This project provides a unified command-line interface (CLI) and orchestration layer to manage multiple Model Context Protocol (MCP) servers. It simplifies setup, configuration, and execution, providing a consistent developer experience across all servers, with powerful features for multi-account management.

## Index

- [Core Concepts](#core-concepts)
- [CLI Architecture (v2)](#cli-architecture-v2)
- [Getting Started](#getting-started-a-detailed-walkthrough)
- [Focused Setups for Specific Workflows](#focused-setups-for-specific-workflows)
- [CLI Command Reference](#cli-command-reference)
- [Advanced Usage: The Alias System](#advanced-usage-the-alias-system)
- [Server Reference](#server-reference)
- [License](#license)

## Core Concepts

- **Unified Management**: A single CLI to rule them all. Start, configure, and manage servers written in any language (Node.js, Go, Python, Dart) without worrying about their individual setup requirements.
- **Centralized Configuration**: All server configurations and API keys are managed in a single `.env` file, making it easy to back up and share your setup.
- **Multi-Account Support**: The powerful alias system allows you to configure a single server (like GitHub) to work with multiple accounts (e.g., "personal" and "work") seamlessly.

## CLI Architecture (v2)

The command-line interface now follows a **Single-Responsibility** layout. Instead of one giant `cli.js`, the code is split into purpose-built modules inside the `cli/` folder:

| Path | Responsibility |
| --- | --- |
| `cli/utils.js` | Generic helper functions (`toPrettyName`, `getServers`, etc.) |
| `cli/install_dependencies.js` | Installs / tidies language-specific dependencies for a server. |
| `cli/constants.js` | Project-wide enums such as `ServerLang`. |
| `cli/runners/` | One file *per language* encapsulating the logic to launch that server type:<br/>• `node_runner.js`<br/>• `go_runner.js`<br/>• `python_runner.js`<br/>• `dart_runner.js` |

The entry point `cli/index.js` now acts as a thin **orchestrator**:
1. Parses CLI arguments with Commander.
2. Discovers a server's language from its `mcp.yml`.
3. Delegates actual execution to the matching runner.

---

## Getting Started: A Detailed Walkthrough

Follow these steps to get the entire environment up and running.

### Step 1: Install Prerequisites

Ensure you have the following software installed on your system:
- **[Node.js](https://nodejs.org/)**: Version 18.x or higher is required to run the orchestrator CLI.
- **[Go](https://go.dev/)**: Version 1.21 or higher for Go-based servers.
- **[Python](https://www.python.org/)**: Version 3.10 or higher for Python-based servers.
- **[Dart](https://dart.dev/get-dart)**: For Dart-based servers.

### Step 2: Install Project Dependencies

Open your terminal in the project's root directory and run the following command. This will install the CLI tool's dependencies.

```bash
npm install
```

### Step 3: Create Your Environment File

This is a critical step for authenticating with various services.

1.  In the project root, create a file named `.env`.
2.  Copy the contents of the template below into your new `.env` file.
3.  Fill in the values for the services you intend to use.

```.env
# -----------------------------------------------------------------------------
# API Keys - Add your personal access tokens here.
# See the "Advanced Usage: The Alias System" section to configure multiple accounts.
# -----------------------------------------------------------------------------

# FIGMA: https://www.figma.com/developers/api#access-tokens
FIGMA_API_KEY=your_figma_personal_access_token

# GITHUB: https://github.com/settings/tokens
GITHUB_PERSONAL_ACCESS_TOKEN=ghp_...

# GITLAB: https://gitlab.com/-/profile/personal_access_tokens
GITLAB_TOKEN=glpat-...

# CLICKUP: https://clickup.com/api/
CLICKUP_API_TOKEN=pk_...

# POSTMAN: https://learning.postman.com/docs/developer/postman-api/authentication/#generate-a-postman-api-key
POSTMAN_API_KEY=PMAK-...
```

### Step 4: Install Server Dependencies

Run the `setup` command to automatically find all servers in the `/servers` directory and install their specific language dependencies (e.g., `npm install`, `go mod tidy`, etc.).

```bash
npm run setup
```

### Step 5: Configure Your MCP Client (Cursor)

To make the servers available in Cursor, you must generate the `mcp.json` configuration file. Run the `setup` command with the `--cursor-setup` flag.

```bash
npm run setup -- --cursor-setup
```
*Note: The `--` is important. It tells npm to pass the `--cursor-setup` flag to our script instead of consuming it.*

### Step 6: Restart Your MCP Client

**You must restart Cursor** for it to recognize the new or updated server configurations. After restarting, your tools will be available in the chat.

---

## Focused Setups for Specific Workflows

For a cleaner and faster experience, you can choose to only set up the servers you need.

### Flutter Essentials

If your primary focus is Flutter development, use the `--flutter-essentials` flag. This will configure Cursor to use only the most relevant servers for that workflow, reducing clutter in the tool selection menu.

```bash
npm run setup -- --cursor-setup --flutter-essentials
```

This will install the following servers:
- Better Fetch
- Compass
- Desktop Commander
- DuckDuckGo Search
- Figma
- Postman
- Flutter (Dart)
- Flutter (Python)
- Sequential Thinking

---

## CLI Command Reference

### `npm run setup`

This command prepares all servers. It has two modes of operation.

**1. Default: Install Dependencies**

Running the command without any flags will iterate through all available servers and install their specific language dependencies.

```bash
npm run setup
```

**2. Cursor Configuration (`--cursor-setup`)**

To generate the `mcp.json` file and runner scripts required by Cursor, use the `--cursor-setup` flag.

```bash
npm run setup -- --cursor-setup
```

### `npm run start <server-id>`

This command builds and runs a specific server by its ID (the directory name in `/servers`).

```bash
npm start figma-node
```

**Skipping Dependency Installation (`--no-install`)**

By default, `npm start` will first ensure a server's dependencies are installed. To skip this step for a faster start, use the `--no-install` flag.

```bash
npm start figma-node -- --no-install
```

### `npm run list`

Lists all available server IDs found in the `/servers` directory.

```bash
npm run list
```

### `npm run test:servers`
Starts each available server sequentially for 15 seconds to verify they can launch without errors. Useful for debugging your setup.

---

## Advanced Usage: The Alias System

This is the most powerful feature of the orchestrator. It allows you to use a single server with multiple identities (e.g., a personal and a work GitHub account).

### How It Works

The system works by looking for specially named variables in your `.env` file. When you create an aliased server, the script appends the alias name to the base key (`BASE_KEY_ALIASNAME`) and uses that value instead.

Let's walk through an example for Figma.

**1. Define Aliased Keys in `.env`**

In your `.env` file, you have your default `FIGMA_API_KEY`. To add a "Qoder" account, you add a second key with `_QODER` appended. The orchestrator will automatically capitalize the alias name.

```.env
# Default Figma key
FIGMA_API_KEY=fig_personal_xxxxxxxxxx

# Aliased key for the 'Qoder' account
FIGMA_API_KEY_QODER=fig_qoder_xxxxxxxxxx
```

**2. Run Setup with an Alias**

Run the `setup` script again, but this time, provide the alias to the `--cursor-setup` command in the format `server-id:alias-name`.

```bash
npm run setup -- --cursor-setup "figma-node:qoder"
```

You can add multiple aliases for multiple servers: `npm run setup -- --cursor-setup "figma-node:qoder,github-go:work"`.

---

## Server Reference

| Server Name | Language | Source Repository |
| :--- | :--- | :--- |
| Qoder - Better Fetch | Node.js | [better-fetch](https://github.com/flutterninja9/better-fetch) |
| Qoder - Clickup | Node.js | [mcp-clickup](https://github.com/mikah13/mcp-clickup) |
| Qoder - Compass | Node.js | [mcp-compass](https://github.com/liuyoshio/mcp-compass) |
| Qoder - Desktop Commander | Node.js | [DesktopCommanderMCP](https://github.com/wonderwhy-er/DesktopCommanderMCP) |
| Qoder - Docker | Python | [mcp-server-docker](https://github.com/ckreiling/mcp-server-docker) |
| Qoder - DuckDuckGo Search | Python | [duckduckgo-mcp-server](https://github.com/nickclyde/duckduckgo-mcp-server) |
| Qoder - Figma | Node.js | [qoder-mcps](https://github.com/figma/figma-developer-mcp) |
| Qoder - Flutter (Dart) | Dart | [mcp_flutter](https://github.com/Arenukvern/mcp_flutter) |
| Qoder - Flutter (Python) | Python | [flutter-mcp](https://github.com/adamsmaka/flutter-mcp) |
| Qoder - Github | Go | [github-mcp-server](https://github.com/github/github-mcp-server) |
| Qoder - Gitlab | Go | [gitlab-mcp](https://gitlab.com/fforster/gitlab-mcp) |
| Qoder - Postman | Node.js | [mcp-server-postman-tool-generation](https://github.com/giovannicocco/mcp-server-postman-tool-generation) |
| Qoder - Sequential Thinking | Node.js | [sequentialthinking](https://github.com/modelcontextprotocol/servers/tree/main/src/sequentialthinking) |

---

## License
This project is licensed under the **BSD 3-Clause License** – see the [LICENSE](LICENSE) file for full details.

> **Migration note:** The historical `cli.js` in the project root has been removed. All entry-points now live in the `cli/` folder (e.g. `node cli/index.js start <id>`). Existing `npm` scripts already point to the new location, so no action is required for day-to-day use.