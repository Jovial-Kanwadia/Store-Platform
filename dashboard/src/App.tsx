import { QueryClient, QueryClientProvider } from "@tanstack/react-query"
import { RouterProvider, createBrowserRouter } from "react-router-dom"

import Dashboard from "@/pages/Dashboard"

const queryClient = new QueryClient()

const router = createBrowserRouter([
  {
    path: "/",
    element: <Dashboard />,
  },
])

function App() {
  return (
    <QueryClientProvider client={queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>
  )
}

export default App
