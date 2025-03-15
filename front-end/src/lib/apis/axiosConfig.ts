import axios from "axios";

export const api = axios.create({
  baseURL: "http://localhost:8080",
  timeout: 3000,
  withCredentials: true, // This allows cookies to be sent and received
});

// Helper for server-side requests that need authentication
export function createAuthHeaders(cookies: any) {
  const sessionToken = cookies.get('session_token');
  const csrfToken = cookies.get('csrf_token');
  
  return {
    Cookie: sessionToken ? `session_token=${sessionToken}` : '',
    'X-CSRF-Token': csrfToken || ''
  };
}
