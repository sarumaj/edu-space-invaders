document.addEventListener("DOMContentLoaded", async function () {
  const go = new Go(); // Defined in wasm_exec.js
  const envVarPrefix = "SPACE_INVADERS_";

  async function envCallback() {
    try {
      const response = await fetch(".env", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({}),
      });
      const data = await response.json();

      // Filter out only the environment variables that start with "SPACE_INVADERS_"
      // Security: This is a very basic way to filter out only the environment variables that are needed
      // This is not a foolproof way to secure the environment variables
      const env = Object.keys(data)
        .filter((key) => key.startsWith(envVarPrefix))
        .reduce((obj, key) => {
          obj[key] = data[key];
          return obj;
        }, {});

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
