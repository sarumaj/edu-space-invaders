async function envCallback(exponentialBackoff = 1) {
  const delayInMs = 2500;

  try {
    if (!apiKey) {
      throw new Error("API key not set");
    }

    const response = await fetch(".env", {
      method: "GET",
      headers: {
        Accept: "application/json",
      },
    });
    const data = await response.json();
    const prefix = data["_prefix"];

    // Filter out only the environment variables that start with the prefix
    const env = Object.keys(data)
      .filter((key) => key.startsWith(prefix))
      .reduce((obj, key) => {
        obj[key] = data[key];
        return obj;
      }, {});

    setTimeout(envCallback, delayInMs);

    globalThis.go_env = env;
  } catch (err) {
    console.error("Error getting env:", err);

    let newExponentialBackoff = exponentialBackoff * 2;
    setTimeout(
      envCallback,
      delayInMs * exponentialBackoff,
      newExponentialBackoff
    );
  }
}

async function getScoreBoard() {
  try {
    const response = await fetch(`/scores.db`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      throw new Error("Error getting scores");
    }

    const text = await response.text();
    return JSON.parse(text.slice(text.indexOf(";") + 1));
  } catch (err) {
    console.error("Error getting scores:", err);
    return [];
  }
}

async function saveScoreBoard(scores) {
  const apiKey = globalThis.go_apiKey;

  try {
    if (!apiKey) {
      throw new Error("API key not set");
    }

    const response = await fetch(`/scores.db`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(scores),
    });

    if (!response.ok) {
      throw new Error("Error saving scores");
    }
  } catch (err) {
    console.error("Error saving scores:", err);
  }
}

async function loadWasm() {
  const go = new Go(); // Defined in wasm_exec.js

  // Initialize the environment variables and api key
  globalThis.go_env = {};

  // Expose the functions to the Go code
  globalThis.go_getScoreBoard = getScoreBoard;
  globalThis.go_saveScoreBoard = saveScoreBoard;

  // Load and instantiate the WebAssembly module
  const wasmModule = await WebAssembly.instantiateStreaming(
    fetch("main.wasm"),
    go.importObject
  );

  // Run the WebAssembly module
  go.run(wasmModule.instance);

  // Update the environment variables with actual values
  await envCallback();
}

window.addEventListener("load", loadWasm());
