let currentState = null; // Tracks the current state ('online' or 'offline')
let awaitingAck = false; // Tracks if we are waiting for an acknowledgment from the main thread

self.onmessage = function (e) {
  switch (e.data.type) {
    case "ack":
      awaitingAck = false; // Received acknowledgment, reset the waiting state
      break;

    case "start":
      checkHealth();
      break;
  }
};

function checkHealth() {
  const delayInMs = 2500;

  if (awaitingAck) {
    // If waiting for acknowledgment, retry after a delay
    setTimeout(checkHealth, delayInMs);
    return;
  }

  fetch("health", {
    method: "GET",
    headers: { Accept: "application/json" },
  })
    .then((response) => {
      if (!response.ok) {
        return response.text().then((text) => {
          throw new Error(text);
        });
      }

      if (currentState !== "online") {
        // Only send message if state has changed
        self.postMessage({ type: "online" });
        currentState = "online";
        awaitingAck = true; // Set flag to wait for acknowledgment
      }

      // Continue health checks with regular delay if online
      setTimeout(checkHealth, delayInMs);
    })
    .catch((error) => {
      if (currentState !== "offline") {
        // Only send message if state has changed
        self.postMessage({ type: "offline" });
        currentState = "offline";
        awaitingAck = true; // Set flag to wait for acknowledgment
      }

      // Use exponential backoff delay when offline
      setTimeout(checkHealth, delayInMs);
    });
}
