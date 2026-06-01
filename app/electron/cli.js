// SiYuan CLI launcher for the in-process (kernel.node) build.
//
// The desktop app embeds the kernel as an N-API module instead of a standalone
// binary, so the upstream cobra CLI (kernel/cli/cmd) is exposed via the
// kernel.node `runCLI` export. This script runs it in a DEDICATED process
// (never the GUI), so stdout and the exit code propagate to the terminal.
//
// Invoke as a Node process, e.g. via the bundled Electron as Node:
//   ELECTRON_RUN_AS_NODE=1 <SiYuan> <resources>/app/electron/cli.js export md --id <id> -w <workspace>
//
// `serve` is intentionally rejected by the kernel (the GUI uses startKernel).

const path = require("path");
const fs = require("fs");

const kernel = require("../native/build/Release/kernel.node");

// Locate the directory that contains the bundled "appearance/langs".
//   packaged: this file is resources/app/electron/cli.js  -> appDir = resources/
//   dev:      this file is <repo>/app/electron/cli.js      -> appDir = <repo>/app
function resolveAppDir() {
    const candidates = [
        path.resolve(__dirname, "..", ".."), // resources/ (packaged)
        path.resolve(__dirname, ".."),       // app/ (dev)
    ];
    for (const dir of candidates) {
        if (fs.existsSync(path.join(dir, "appearance", "langs"))) {
            return dir;
        }
    }
    return candidates[0];
}

const rc = kernel.runCLI(process.argv.slice(2), resolveAppDir());
process.exit(rc);
