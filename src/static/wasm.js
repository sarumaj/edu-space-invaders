const go = new Go(); // Defined in wasm_exec.js

// Fetch the environment variable from the server
fetch(".env", {
  method: "POST",
  headers: {
    "Content-Type": "application/json",
  },
  body: JSON.stringify({}),
})
  .then((response) => response.json())
  .then((env) => {
    // Set the environment variables in the global context for WASM to access
    globalThis.go_env = env;

    // Fetch and instantiate the WebAssembly module
    return WebAssembly.instantiateStreaming(
      fetch("main.wasm"),
      go.importObject
    );
  })
  .then((result) => {
    go.run(result.instance);
  })
  .catch((err) => {
    console.error("Error instantiating WASM module:", err);
  });
