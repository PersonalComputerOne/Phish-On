# extension

### Prerequisites

- [Bun](https://bun.sh)

```bash
# Install Bun (if not already installed)
https://bun.sh/docs/installation
```

### 📦 Install dependencies

```bash
bun i
```

### 🛠 Chrome Extension Installation

1. **Build the extension**

   ```bash
   bun run build
   ```

2. Open Chrome and navigate to:

   ```
   chrome://extensions
   ```

3. Enable **Developer Mode**:

   - Toggle the switch in the top-right corner

4. **Load Unpacked Extension**:
   - Click the "Load unpacked" button
   - Select the `dist` folder from your project directory
   - Click "Open"
