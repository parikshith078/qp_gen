import { api, createAuthHeaders } from '$lib/apis/axiosConfig';
import { type ApiResponse, type UserModel } from '$lib/apis/types';
import type { Handle } from '@sveltejs/kit';
import { redirect } from '@sveltejs/kit';

// Define public routes that don't require authentication
const publicRoutes = ['/', '/login', '/register', '/about'];

export const handle: Handle = async ({ event, resolve }) => {
	// Check if the current path is a public route
	const isPublicRoute = publicRoutes.some(route => 
		event.url.pathname === route || event.url.pathname === `${route}/`
	);

	// get cookies from browser
	const userid = event.cookies.get('userid');

	if (!userid) {
		// if there is no session and not a public route, redirect to login
		if (!isPublicRoute) {
			throw redirect(303, '/login');
		}
		// otherwise load page as normal
		return await resolve(event);
	}

	// find the user based on the session
	try {
		const headers = createAuthHeaders(event.cookies);
		const { data: res } = await api.get<ApiResponse<UserModel>>(`/user/${userid}`, { headers });
		
		if (res.data) {
			event.locals.user = res.data;
		} else {
			// If API returns success but no user data, clear cookies and redirect
			if (!isPublicRoute) {
				event.cookies.delete('session_token', { path: '/' });
				event.cookies.delete('csrf_token', { path: '/' });
				event.cookies.delete('userid', { path: '/' });
				throw redirect(303, '/login');
			}
		}
	} catch (error) {
		// If API call fails, clear cookies and redirect if not public route
		if (!isPublicRoute) {
			event.cookies.delete('session_token', { path: '/' });
			event.cookies.delete('csrf_token', { path: '/' });
			event.cookies.delete('userid', { path: '/' });
			throw redirect(303, '/login');
		}
	}

	// load page as normal
	return await resolve(event);
};
