import "leaflet/dist/leaflet.css";
import React from "react";
import ReactDOM from "react-dom/client";
import App from "./App";
import "./index.css";
import reportWebVitals from "./reportWebVitals";
import "./styles/design-tokens.css";
import { initializeTraceFetch } from "./utils/traceFetch";

initializeTraceFetch();

// posthog.init(import.meta.env.VITE_PUBLIC_POSTHOG_KEY || "dummy_key", {
//   api_host:
//     import.meta.env.VITE_PUBLIC_POSTHOG_HOST || "https://us.i.posthog.com",
//   person_profiles: "always",
// });

const root = ReactDOM.createRoot(
  document.getElementById("root") as HTMLElement,
);
root.render(
  <React.StrictMode>
    {/* <PostHogProvider client={posthog}> */}
      <App />
    {/* </PostHogProvider> */}
  </React.StrictMode>,
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
