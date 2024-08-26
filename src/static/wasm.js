async function loadWasm() {
  const go = new Go(); // Defined in wasm_exec.js

  // Show the loading overlay and apply the background effects
  const overlay = document.getElementById("loadingOverlay");
  if (overlay.classList.contains("hidden")) {
    overlay.classList.remove("hidden");
    overlay.classList.add("active");
  }

  // Initialize the environment variables
  globalThis.go_env = {};

  // Load and instantiate the WebAssembly module
  const wasmModule = await WebAssembly.instantiateStreaming(
    fetch("main.wasm"),
    go.importObject
  );

  // Change overlay text to "Running..."
  const overlayText = document.getElementById("loadingMessage");
  overlayText.innerText = "Game loaded. Running...";

  // Introduce a small delay to ensure the DOM is updated before running the Wasm module
  await new Promise((resolve) => setTimeout(resolve, 100)); // 100ms delay

  // Run the WebAssembly module
  go.run(wasmModule.instance);

  // Hide the loading overlay and remove the background effects
  overlay.classList.add("hidden");
  overlay.classList.remove("active");
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
