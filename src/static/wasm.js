async function envCallback() {
  try {
    const response = await fetch(".env", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({}),
    });
    const data = await response.json();

    // Filter out only the environment variables that start with "SPACE_INVADERS_"
    const env = Object.keys(data)
      .filter((key) => key.startsWith("SPACE_INVADERS_"))
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

async function loadWasm() {
  const go = new Go(); // Defined in wasm_exec.js

  globalThis.go_env = await envCallback();

  const wasmModule = await WebAssembly.instantiateStreaming(
    fetch("main.wasm"),
    go.importObject
  );

  go.run(wasmModule.instance);
}

window.addEventListener("load", loadWasm());
