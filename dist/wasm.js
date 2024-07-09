document.addEventListener("DOMContentLoaded", async function () {
  const go = new Go(); // Defined in wasm_exec.js

  async function envCallback() {
    try {
      const response = await fetch(".env", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({}),
      });
      const env = await response.json();
      return env;
    } catch (err) {
      console.error("Error getting env:", err);
      return {};
    }
  }

  globalThis.go_env = await envCallback();
  WebAssembly.instantiateStreaming(fetch("main.wasm"), go.importObject)
    .then((result) => {
      go.run(result.instance);
    })
    .catch((err) => {
      console.error("Error instantiating WASM module:", err);
    });
});
