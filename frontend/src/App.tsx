// App.tsx
import { Suspense } from "react";
import { useRoutes } from "react-router-dom";
import routes from "~react-pages";

export default function App() {
  const element = useRoutes(routes);
  return <Suspense fallback={<p>Loading...</p>}>{element}</Suspense>;
}
