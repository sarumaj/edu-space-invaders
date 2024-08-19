async function loadWasm() {
  const go = new Go(); // Defined in wasm_exec.js

  // Initialize the environment variables
  globalThis.go_env = {};

  // Load and instantiate the WebAssembly module
  const wasmModule = await WebAssembly.instantiateStreaming(
    fetch("main.wasm"),
    go.importObject
  );

  // Run the WebAssembly module
  go.run(wasmModule.instance);
}

window.addEventListener("load", loadWasm());

if ("serviceWorker" in navigator) {
  navigator.serviceWorker.register("service-worker.js").then(
    (registration) => {
      console.log(
        "ServiceWorker registration successful with scope: ",
        registration.scope
      );
    },
    (error) => {
      console.log("ServiceWorker registration failed: ", error);
    }
  );
}
