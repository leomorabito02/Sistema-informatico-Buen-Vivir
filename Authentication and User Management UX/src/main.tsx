
  import { createRoot } from "react-dom/client";
  import { PrimeReactProvider } from 'primereact/api';
  import 'primereact/resources/themes/lara-light-green/theme.css';
  import 'primereact/resources/primereact.min.css';
  import 'primeicons/primeicons.css';
  import App from "./app/App.tsx";
  import "./styles/index.css";

  createRoot(document.getElementById("root")!).render(
    <PrimeReactProvider>
      <App />
    </PrimeReactProvider>
  );
  