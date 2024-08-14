// The key will be set in the Go code
var apiKey = "";

async function envCallback() {
  try {
    if (!apiKey) {
      throw new Error("API key not set");
    }

    const response = await fetch(".env", {
      method: "GET",
      headers: {
        Accept: "application/json",
        Authorization: `Bearer ${apiKey}`,
      },
    });
    const data = await response.json();
    const prefix = data["_prefix"];

    // Filter out only the environment variables that start with "SPACE_INVADERS_"
    const env = Object.keys(data)
      .filter((key) => key.startsWith(prefix))
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
  try {
    if (!apiKey) {
      throw new Error("API key not set");
    }

    const response = await fetch(`/scores.db`, {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        Authorization: `Bearer ${apiKey}`,
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

  // Pass the environment variables to the Go code
  globalThis.go_env = await envCallback();
  globalThis.go_scoreBoard = await getScoreBoard();

  // Expose the functions to the Go code
  globalThis.go_saveScoreBoard = saveScoreBoard;

  const wasmModule = await WebAssembly.instantiateStreaming(
    fetch("main.wasm"),
    go.importObject
  );

  go.run(wasmModule.instance);
}

window.addEventListener("load", loadWasm());
