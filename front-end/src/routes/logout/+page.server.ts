import { api, createAuthHeaders } from '$lib/apis/axiosConfig';
import { error, redirect } from '@sveltejs/kit';
import type { AxiosError } from 'axios';
import type { ApiResponse } from '$lib/apis/types';
import type { PageServerLoad } from '../register/$types';

export const load = (async ({ cookies }) => {
	const headers = createAuthHeaders(cookies);

	// Send the request with cookies
	api.post<ApiResponse>('/logout', {}, { headers });

	// Delete cookies
	cookies.delete('session_token', { path: '/' });
	cookies.delete('csrf_token', { path: '/' });
	cookies.delete('userid', { path: '/' });

	throw redirect(303, '/login');
}) satisfies PageServerLoad;
