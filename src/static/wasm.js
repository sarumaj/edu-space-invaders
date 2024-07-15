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

async function onResize(redrawFunc) {
  const document = window.document;
  const canvas = document.getElementById("gameCanvas");
  const ctx = canvas.getContext("2d");

  // Get current canvas width and height
  const width = canvas.width;
  const height = canvas.height;

  // Get image data from canvas
  const data = ctx.getImageData(0, 0, width, height);

  // Get the dimensions of the container
  const innerWidth = canvas.clientWidth;
  const innerHeight = canvas.clientHeight;

  // Set new canvas width and height
  canvas.width = innerWidth;
  canvas.height = innerHeight;

  // Put the image data back to the canvas
  ctx.putImageData(data, 0, 0);

  if (redrawFunc) {
    console.log("Redrawing...");
    redrawFunc();
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

  // Initialize audioEnabled from Go
  const isAudioEnabledFunc = window.isAudioEnabled; // Ensure we have a reference to the function
  const toggleAudioFunc = window.toggleAudio; // Ensure we have a reference to the function

  let audioEnabled = await isAudioEnabledFunc(); // Ensure that isAudioEnabled is awaited and set
  const audioIcon = document.getElementById("audioIcon");
  if (audioEnabled) {
    audioIcon.classList.remove("fa-volume-mute");
    audioIcon.classList.add("fa-volume-up");
  } else {
    audioIcon.classList.remove("fa-volume-up");
    audioIcon.classList.add("fa-volume-mute");
  }

  window.toggleAudio = async function () {
    await toggleAudioFunc(); // Call the Go function to toggle audio
    audioEnabled = await isAudioEnabledFunc(); // Get the updated audio state
    if (audioEnabled) {
      audioIcon.classList.remove("fa-volume-mute");
      audioIcon.classList.add("fa-volume-up");
    } else {
      audioIcon.classList.remove("fa-volume-up");
      audioIcon.classList.add("fa-volume-mute");
    }
  };

  const audioToggleBtn = document.getElementById("audioToggle");
  audioToggleBtn.addEventListener("click", window.toggleAudio);
  audioToggleBtn.addEventListener("touchend", function (event) {
    event.preventDefault(); // Prevent mouse event from also being triggered
    toggleAudio();
  });

  const refreshBtn = document.getElementById("refreshButton");
  function animateButton() {
    return new Promise((resolve) => {
      refreshBtn.classList.add("animated-click");
      refreshBtn.addEventListener(
        "transitionend",
        () => {
          refreshBtn.classList.remove("animated-click");
          refreshBtn.classList.add("animated-click-end");
          resolve();
        },
        { once: true }
      );
    });
  }
  refreshBtn.addEventListener("click", () => {
    animateButton().then(() => {
      location.reload(); // Reload after the animation completes
    });
  });
  refreshBtn.addEventListener("touchend", function (event) {
    event.preventDefault(); // Prevent mouse event from also being triggered
    animateButton().then(() => {
      location.reload(); // Reload after the animation completes
    });
  });

  const redrawFunc = window.redrawContent;
  window.addEventListener("resize", () => {
    requestAnimationFrame(onResize, redrawFunc);
  });

  await onResize(redrawFunc);
}

window.addEventListener("load", loadWasm());
