document.addEventListener("DOMContentLoaded", async function () {
  const go = new Go(); // Defined in wasm_exec.js

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

  globalThis.go_env = await envCallback();

  const wasmModule = await WebAssembly.instantiateStreaming(
    fetch("main.wasm"),
    go.importObject
  );
  go.run(wasmModule.instance);

  // Initialize audioEnabled from Go
  const isAudioEnabledFunc = window.isAudioEnabled; // Ensure we have a reference to the function
  const toggleAudioFunc = window.toggleAudio; // Ensure we have a reference to the function

  let audioEnabled = await isAudioEnabledFunc(); // Ensure that isAudioEnabled is awaited and set
  const audioIcon = document.getElementById("audioIcon");
  if (audioEnabled) {
    audioIcon.className = "fas fa-volume-up";
  } else {
    audioIcon.className = "fas fa-volume-mute";
  }

  window.toggleAudio = async function () {
    await toggleAudioFunc(); // Call the Go function to toggle audio
    audioEnabled = await isAudioEnabledFunc(); // Get the updated audio state
    if (audioEnabled) {
      audioIcon.className = "fas fa-volume-up";
    } else {
      audioIcon.className = "fas fa-volume-mute";
    }
  };

  audioToggleBtn = document.getElementById("audioToggle");
  audioToggleBtn.addEventListener("click", window.toggleAudio);
  audioToggleBtn.addEventListener("touchend", function (event) {
    event.preventDefault(); // Prevent mouse event from also being triggered
    toggleAudio();
  });
});
